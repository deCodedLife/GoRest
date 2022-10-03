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
	var API RestApi

	API.Path = s.Table

	API.Handler = func(w http.ResponseWriter, r *http.Request) {

	}
	API.Method = http.MethodGet
	Handlers = append(Handlers, API)

	API.Path = s.Table
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		var userRequest map[string]interface{}

		defer func() {
			recover()
		}()

		err := json.NewDecoder(r.Body).Decode(&userRequest)
		HandleError(err, CustomError{}.WebError(w, 401, err))

		err = s.ValidateParams(userRequest)
		HandleError(err, CustomError{}.WebError(w, 401, err))

		id, err := s.INSERT(userRequest)
		HandleError(err, CustomError{}.WebError(w, 501, err))

		SendData(w, 200, id)
	}
	API.Method = http.MethodPost
	Handlers = append(Handlers, API)

	API.Path = s.Table + "/{id}"
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		defer recover()

		conditionerID, err := strconv.Atoi(vars["id"])
		HandleError(err, CustomError{}.WebError(w, http.StatusForbidden, err))

		SendData(w, http.StatusOK, fmt.Sprintf("[%d] was deleted", conditionerID))
	}
	API.Method = http.MethodDelete
	Handlers = append(Handlers, API)

	API.Path = s.Table
	API.Handler = func(w http.ResponseWriter, r *http.Request) {
		SendData(w, http.StatusOK, fmt.Sprintf("[%s] PUT method works fine", s.Table))
	}
	API.Method = http.MethodPut
	Handlers = append(Handlers, API)
}

func Construct() {
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

		HandleRest(dbSchema)
		dbSchema.InitTable()
	}

	r := mux.NewRouter()

	for _, api := range Handlers {
		r.HandleFunc("/"+api.Path, api.Handler).Methods(api.Method)
	}

	err = http.ListenAndServe(":80", r)
	HandleError(err, CustomError{}.Unxepected(err))
}
