package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/tuken/nix/graph"
	"github.com/tuken/nix/logger"
	"github.com/tuken/nix/middleware"
	"github.com/tuken/nix/pipeline"
	"github.com/vektah/gqlparser/v2/ast"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

func main() {

	logDirectory := os.Getenv("LOG_DIRECTORY")
	if logDirectory == "" {
		logDirectory = "./"
	}

	listenPort := os.Getenv("LISTEN_PORT")
	if listenPort == "" {
		listenPort = "8080"
	}

	logLevel := gormlog.Error

	sqlLogLevel := os.Getenv("SQL_LOG_LEVEL")
	switch sqlLogLevel {

	case "warn":
		logLevel = gormlog.Warn

	case "info":
		logLevel = gormlog.Info
	}

	mainLog := logger.NewLogger(logDirectory + "/nix.log")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_bin&parseTime=True&loc=Asia%%2FTokyo",
		os.Getenv("DB_USER"),
		url.QueryEscape(os.Getenv("DB_PASSWORD")),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"))

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: mainLog})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	db.Logger = mainLog.LogMode(logLevel)

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{}}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	srv.AroundOperations(middleware.LoggingMiddleware)

	// ヘルスチェックエンドポイント
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {})

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	pre := &pipeline.Preprocessor{DB: db, Log: mainLog, SQLLogLevel: logLevel}
	http.Handle("/query", pre.Pipeline(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", listenPort)
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
