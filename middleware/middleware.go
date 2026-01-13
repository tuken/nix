package middleware

import (
	"context"
	"log"

	"github.com/99designs/gqlgen/graphql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type contextKey struct{ string }

var (
	DBKey = contextKey{"database"}
)

func DatabaseMiddleware(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {

	log.Println("OperationMiddleware: called")

	dsn := "root:secualpass@tcp(localhost:3306)/agrimo?charset=utf8mb4&collation=utf8mb4_bin&parseTime=True&loc=Asia%2FTokyo"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	newCtx := context.WithValue(ctx, DBKey, db)

	return next(newCtx)
}

func GetDatabase(ctx context.Context) *gorm.DB {

	db, ok := ctx.Value(DBKey).(*gorm.DB)
	if !ok {
		log.Fatalf("failed to get database from context")
	}

	return db
}
