package graphqltransport

import (
	"net/http"

	"github.com/graphql-go/graphql"

	"morent-backend/internal/handlers"
	"morent-backend/internal/modules/transport/http/common"
)

func Register(mux *http.ServeMux, schema graphql.Schema) {
	mux.HandleFunc("/graphql", common.WrapCORS(handlers.NewGraphQLHandler(schema)))
}
