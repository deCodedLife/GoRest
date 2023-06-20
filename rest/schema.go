package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	. "github.com/deCodedLife/gorest/database"
	. "github.com/deCodedLife/gorest/tool"
	"github.com/gorilla/mux"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type SchemaStructure struct {
	Table  string        `json:"table"`
	Schema []SchemaParam `json:"schema"`
}

var SCHEMAS []Schema

func ParamsToQuery(s Schema, query url.Values) map[string]interface{} {
	var uriParams = make(map[string]interface{})
	for _, param := range s.Params {
		if param.Type == "" {
			continue
		}
		var valueExists bool
		for variable := range query {
			if variable == param.Article {
				value := query.Get(variable)
				if value == "" {
					break
				}
				valueExists = true
				break
			}
		}
		if valueExists == false {
			continue
		}
		uriParams[param.Article] = query.Get(param.Article)
	}
	return uriParams
}

func ExtendObjects() {
	for _, s := range SCHEMAS {
		Handlers = append(Handlers, RestApi{
			Path:   "serialized-" + s.Title,
			Method: http.MethodGet,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					recover()
				}()
				data, err := s.SELECT(ParamsToQuery(s, r.URL.Query()))
				if len(data) == 0 {
					SendData(w, http.StatusOK, nil)
					return
				}
				object := data[0]
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				for _, param := range s.Params {
					if param.TakeFrom == "" && param.Join == "" {
						continue
					}

					relatedObject := strings.Split(param.TakeFrom, "/")[0]
					relatedParam := "id"

					if param.Join != "" {
						relatedObject = strings.Split(param.Join, "/")[0]
						relatedParam = strings.Split(param.Join, "/")[1]
					}

					for _, scheme := range SCHEMAS {
						if scheme.Table != relatedObject {
							continue
						}

						request := make(map[string]interface{})
						request[relatedParam] = object[param.Article]
						relatedList, err := scheme.SELECT(request)

						if len(relatedList) == 0 {
							object[param.Article] = nil
							continue
						}

						HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))
						object[param.Article] = relatedList[0]
					}
				}

				SendData(w, http.StatusOK, object)
			},
		})
	}
}

func HandleRest(s Schema) {
	Handlers = append(Handlers, RestApi{
		Path:   s.Table + "/schema",
		Method: http.MethodGet,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			returnStructure := SchemaStructure{
				Table:  s.Title,
				Schema: s.Params,
			}
			SendData(w, 200, returnStructure)
		},
	})

	if s.ContainsMethod("GET") {
		Handlers = append(Handlers, RestApi{
			Path:   s.Table,
			Method: http.MethodGet,
			Handler: func(w http.ResponseWriter, r *http.Request) {
				defer func() {
					recover()
				}()
				data, err := s.SELECT(ParamsToQuery(s, r.URL.Query()))
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
				HandleError(err, CustomError{}.WebError(
					w,
					http.StatusNotAcceptable,
					errors.New("not allowed"),
				))
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
				HandleError(err, CustomError{}.WebError(
					w,
					http.StatusNotAcceptable,
					errors.New("not allowed"),
				))

				err = json.NewDecoder(r.Body).Decode(&userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusNotAcceptable, err))

				rowsAffected, err := s.UPDATE(id, userRequest)
				HandleError(err, CustomError{}.WebError(w, http.StatusInternalServerError, err))

				SendData(w, http.StatusOK, rowsAffected)
			},
		})
	}
}

func GetSchemas() ([]Schema, error) {

	var schemasList []Schema

	filesList, err := os.ReadDir(SchemaDir)

	if err != nil {
		return nil, err
	}

	for _, file := range filesList {

		var dbSchema Schema

		byteData, err := os.ReadFile(filepath.Join(SchemaDir, file.Name()))

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(byteData, &dbSchema)

		if err != nil {
			return nil, err
		}

		schemasList = append(schemasList, dbSchema)
	}

	SCHEMAS = schemasList
	return schemasList, nil
}

func Construct() []RestApi {
	DBConfig.Init()
	InitDatabase()

	schemasList, err := GetSchemas()
	HandleError(err, CustomError{}.Unexpected(err))

	for _, schema := range schemasList {
		HandleRest(schema)
		schema.InitTable()
	}

	ExtendObjects()

	return Handlers
}
