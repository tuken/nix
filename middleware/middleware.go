package middleware

import (
	"context"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/tuken/nix/pipeline"
)

func LoggingMiddleware(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {

	l := pipeline.MustLogger(ctx)

	start := time.Now()

	oc := graphql.GetOperationContext(ctx)

	l.Infow("GraphQL Query Started", "operation", oc.OperationName, "query", oc.RawQuery, "variables", oc.Variables)

	response := next(ctx)

	duration := time.Since(start)

	defer func() {
		l.Infow("GraphQL Query Completed", "operation", oc.OperationName, "duration", duration)
	}()

	return response
}
