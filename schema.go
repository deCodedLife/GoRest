package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
)

type SchemaParam struct {
	Title string `json:"title"`
	Article string `json:"article"`
	Type string `json:"type"`
	Null string `json:"null"`
	Default string `json:"default"`
}

type Schema struct {
	Title string `json:"title"`
	Table string `json:"table"`
	Params []SchemaParam `json:"params"`
}

func (s Schema) Parse() {

	var API RestApi

	API.Path = s.Table

	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		SendData(w, http.StatusOK, fmt.Sprintf("[%s] GET method works fine!", s.Table))
	}
	API.Method = http.MethodGet
	handlers = append(handlers, API)

	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		SendData(w, http.StatusOK, fmt.Sprintf("[%s] POST method works fine", s.Table))
	}
	API.Method = http.MethodPost
	handlers = append(handlers, API)

	API.Path = s.Table + "/{id}"
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		defer recover()

		conditionerID, err := strconv.Atoi( vars["id"] )
		HandleError(err, CustomError{}.WebError(w, http.StatusForbidden, err))

		SendData(w, http.StatusOK, fmt.Sprintf("[%d] was deleted", conditionerID))
	}
	API.Method = http.MethodDelete
	handlers = append(handlers, API)

	API.Path = s.Table
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		SendData(w, http.StatusOK, fmt.Sprintf("[%s] PUT method works fine", s.Table))
	}
	API.Method = http.MethodPut
	handlers = append(handlers, API)

}

func construct() {

	DBConfig.init()
	InitDatabase()

	filesList, err := ioutil.ReadDir(SchemaDir)
	HandleError(err, CustomError{}.Unxepected(err))

	for _, file := range filesList {
		var dbSchema Schema

		byteData, err := ioutil.ReadFile(filepath.Join(SchemaDir, file.Name()))
		HandleError(err, CustomError{}.Unxepected(err))

		err = json.Unmarshal(byteData, &dbSchema)
		HandleError(err, CustomError{}.Unxepected(err))

		dbSchema.Parse()
		dbSchema.InitTable()
	}

	r := mux.NewRouter()

	for _, api := range handlers {
		log.Println(api.Method + " /" + api.Path)
		r.HandleFunc("/" + api.Path, api.Handler).Methods(api.Method)
	}

	err = http.ListenAndServe(":80", r)
	HandleError(err, CustomError{}.Unxepected(err))
}
