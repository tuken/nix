package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/oklog/ulid/v2"
	"github.com/tuken/nix/logger"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
)

type contextKey struct{ string }

var (
	DBKey     = contextKey{"database"}
	LoggerKey = contextKey{"logger"}
)

// QueryLogger handles query logging with DB and Logger access
type QueryLogger struct {
	DB          *gorm.DB
	Log         logger.Logger
	SQLLogLevel gormlog.LogLevel
}

// Middleware is the middleware function that logs queries and sets context values
func (ql *QueryLogger) Middleware(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {

	requestID := ulid.Make().String()

	db := ql.DB.WithContext(ctx)

	log := ql.Log.With("rid", requestID)
	db.Logger = log.LogMode(ql.SQLLogLevel)

	ctx = context.WithValue(ctx, DBKey, db)
	ctx = context.WithValue(ctx, LoggerKey, log)

	start := time.Now()

	oc := graphql.GetOperationContext(ctx)

	log.Infow("GraphQL Query Started", "operation", oc.OperationName, "query", oc.RawQuery, "variables", oc.Variables)

	response := next(ctx)

	duration := time.Since(start)

	log.Infow("GraphQL Query Completed", "operation", oc.OperationName, "duration", duration)

	return response
}

func MustDB(ctx context.Context) *gorm.DB {

	if db, ok := ctx.Value(DBKey).(*gorm.DB); ok {
		return db
	}

	panic("missing db value in context")
}

func MustLogger(ctx context.Context) logger.Logger {

	if log, ok := ctx.Value(LoggerKey).(logger.Logger); ok {
		return log
	}

	panic("missing logger value in context")
}
