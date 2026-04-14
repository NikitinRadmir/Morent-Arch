package handlers

import (
	"net/http"

	gqlhandler "github.com/graphql-go/handler"
	"github.com/graphql-go/graphql"
)

// оздаёт http.HandlerFunc для /graphql на основе переданной схемы.
func NewGraphQLHandler(schema graphql.Schema) http.HandlerFunc {
	h := gqlhandler.New(&gqlhandler.Config{
		Schema:   &schema,
		Pretty:   true,
		GraphiQL: true,
	})

	return func(w http.ResponseWriter, r *http.Request) {
		// handler сам пишет JSON-ответ
		h.ServeHTTP(w, r)
	}
}

