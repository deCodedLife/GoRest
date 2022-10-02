package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

type SchemaParam struct {
	Title   string `json:"title"`
	Article string `json:"article"`
	Type    string `json:"type"`
	Null    string `json:"null"`
	Default string `json:"default"`
}

type Schema struct {
	Title  string        `json:"title"`
	Table  string        `json:"table"`
	Params []SchemaParam `json:"params"`
}

func (s SchemaParam) IsNumeric() bool {
	if len(strings.Split(s.Type, "bit")) > 1 {
		return true
	}
	if len(strings.Split(s.Type, "bool")) > 1 {
		return true
	}
	if len(strings.Split(s.Type, "int")) > 1 {
		return true
	}
	if len(strings.Split(s.Type, "float")) > 1 {
		return true
	}
	if len(strings.Split(s.Type, "double")) > 1 {
		return true
	}
	if len(strings.Split(s.Type, "dec")) > 1 {
		return true
	}

	return false
}

func (s Schema) Parse() {
	var API RestApi

	API.Path = s.Table

	API.Handler = func(w http.ResponseWriter, r *http.Request) {

	}
	API.Method = http.MethodGet
	handlers = append(handlers, API)

	API.Path = s.Table
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		var userRequest map[string]interface{}

		defer func() {
			recover()

			//HandleError(errors.New(err.(string)), CustomError{}.Unxepected(errors.New(err.(string))))
		}()

		err := json.NewDecoder(r.Body).Decode(&userRequest)
		HandleError(err, CustomError{}.WebError(w, 401, err))

		for _, param := range s.Params {

			if param.Null != "NO" || strings.ToLower(param.Article) == "id" {
				continue
			}

			if param.Default != "" {
				continue
			}

			if userRequest[param.Article] == nil {
				SendData(w, 401, fmt.Sprintf("%s is required", param.Article))
				return
			}

		}

		id, err := s.INSERT(userRequest)
		HandleError(err, CustomError{}.WebError(w, 501, err))

		SendData(w, 200, id)
	}
	API.Method = http.MethodPost
	handlers = append(handlers, API)

	API.Path = s.Table + "/{id}"
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		defer recover()

		conditionerID, err := strconv.Atoi(vars["id"])
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
	DBConfig.Init()
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
		r.HandleFunc("/"+api.Path, api.Handler).Methods(api.Method)
	}

	err = http.ListenAndServe(":80", r)
	HandleError(err, CustomError{}.Unxepected(err))
}
