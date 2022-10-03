package rest

import (
	"net/http"
)

const SchemaDir = "schema"

var Handlers []RestApi

type RestApi struct {
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
	Method  string
}
