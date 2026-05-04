package scalars

import (
	"fmt"
	"io"
	"strconv"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
)

func MarshalUUID(u uuid.UUID) graphql.Marshaler {
	if u == uuid.Nil {
		return graphql.Null
	}
	return graphql.WriterFunc(func(w io.Writer) {
		_, _ = io.WriteString(w, strconv.Quote(u.String()))
	})
}

func UnmarshalUUID(v interface{}) (uuid.UUID, error) {
	switch v := v.(type) {
	case string:
		return uuid.Parse(v)
	case *string:
		if v == nil {
			return uuid.Nil, nil
		}
		return uuid.Parse(*v)
	case []byte:
		return uuid.Parse(string(v))
	default:
		return uuid.Nil, fmt.Errorf("invalid type %T for UUID", v)
	}
}
