package main

import (
	"github.com/gorilla/mux"
	"net/http"

	. "ouralbum/gorest/rest"
	. "ouralbum/gorest/tool"
)

func main() {
	Handlers := Construct()

	r := mux.NewRouter()

	for _, api := range Handlers {
		r.HandleFunc("/"+api.Path, api.Handler).Methods(api.Method)
	}

	err := http.ListenAndServe(":80", r)
	HandleError(err, CustomError{}.Unexpected(err))
}
