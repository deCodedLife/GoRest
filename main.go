package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const SchemaDir = "schema"

var handlers []RestApi

type RestApi struct {
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
	Method  string
}

type CustomError struct {
	StatusCode int32 `json:"status_code"`
	Handler    func()
}

type Response struct {
	StatusCode int32       `json:"status_code"`
	Data       interface{} `json:"data"`
}

func (c CustomError) Unxepected(err error) CustomError {
	return CustomError{
		500,
		func() {
			PrintLog("error", "main", err.Error())
			panic(err.Error())
		}}
}

func (c CustomError) WebError(w http.ResponseWriter, s int32, err error) CustomError {

	return CustomError{
		s,
		func() {
			SendData(w, s, err.Error())
			panic(err.Error())
		},
	}
}

func PrintLog(t string, s string, d interface{}) {
	log.Println(fmt.Sprintf("[%s] %s: \"%s\"", t, s, d.(string)))
}

func SendData(w http.ResponseWriter, s int32, d interface{}) {

	w.Header().Set("Content-Type", "application/json")

	var response Response
	response.StatusCode = s
	response.Data = d

	err := json.NewEncoder(w).Encode(response)
	HandleError(err, CustomError{}.Unxepected(err))
}

func HandleError(err error, d CustomError) {
	if err != nil {
		d.Handler()
	}
}

func main() {
	construct()
}
