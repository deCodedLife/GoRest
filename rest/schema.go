package rest

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"

	. "backend/database"
	. "backend/tool"
)

func HandleRest(s Schema) {
	if s.ContainsMethod("GET") {
		Handlers = append(Handlers, RestApi{
			Path:   s.Table,
			Method: http.MethodGet,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var userRequest map[string]interface{}

				defer func() {
					recover()
				}()

				err := json.NewDecoder(r.Body).Decode(&userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				data, err := s.SELECT(userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				SendData(w, 200, data)
			},
		})
	}

	if s.ContainsMethod("POST") {
		Handlers = append(Handlers, RestApi{
			Path:   s.Table,
			Method: http.MethodPost,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var userRequest map[string]interface{}

				defer func() {
					recover()
				}()

				err := json.NewDecoder(r.Body).Decode(&userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				err = s.ValidateParams(userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				id, err := s.INSERT(userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				SendData(w, 200, id)
			},
		})
	}

	if s.ContainsMethod("DELETE") {
		Handlers = append(Handlers, RestApi{
			Path:   fmt.Sprintf("%s/{id}", s.Table),
			Method: http.MethodDelete,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)

				defer func() {
					recover()
				}()

				id, err := strconv.Atoi(vars["id"])
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				rowsAffected, err := s.DELETE(id)
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				SendData(w, http.StatusOK, rowsAffected)
			},
		})
	}

	if s.ContainsMethod("PUT") {
		Handlers = append(Handlers, RestApi{
			Path:   fmt.Sprintf("%s/{id}", s.Table),
			Method: http.MethodPut,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				var userRequest map[string]interface{}
				vars := mux.Vars(r)

				defer func() {
					recover()
				}()

				id, err := strconv.Atoi(vars["id"])
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				json.NewDecoder(r.Body).Decode(userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				rowsAffected, err := s.UPDATE(id, userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				SendData(w, http.StatusOK, rowsAffected)
			},
		})
	}
}

func Construct() {
	DBConfig.Init()
	InitDatabase()

	filesList, err := ioutil.ReadDir(SchemaDir)
	HandleError(err, CustomError{}.Unexpected(err))

	for _, file := range filesList {
		var dbSchema Schema

		byteData, err := ioutil.ReadFile(filepath.Join(SchemaDir, file.Name()))
		HandleError(err, CustomError{}.Unexpected(err))

		err = json.Unmarshal(byteData, &dbSchema)
		HandleError(err, CustomError{}.Unexpected(err))

		HandleRest(dbSchema)
		dbSchema.InitTable()
	}

	r := mux.NewRouter()

	for _, api := range Handlers {
		r.HandleFunc("/"+api.Path, api.Handler).Methods(api.Method)
	}

	err = http.ListenAndServe(":80", r)
	HandleError(err, CustomError{}.Unexpected(err))
}
