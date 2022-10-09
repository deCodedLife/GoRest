package rest

import (
	"net/http"
)

const SchemaDir = "schema"

var Handlers []RestApi

type Query struct {
	Name  string
	Param string
}

type RestApi struct {
	Method  string
	Path    string
	Queries []Query
	Handler func(w http.ResponseWriter, r *http.Request)
}
