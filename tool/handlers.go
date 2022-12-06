package tool

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const DBConfigFile = "dbSettings.json"

func (c CustomError) Unexpected(err error) CustomError {
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers:", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.Header().Set("Content-Type", "application/json")

	var response Response
	response.StatusCode = s
	response.Data = d

	jsonData, err := json.Marshal(response)
	HandleError(err, CustomError{}.Unexpected(err))

	_, err = w.Write(jsonData)
	HandleError(err, CustomError{}.Unexpected(err))
}

func HandleError(err error, d CustomError) {
	if err != nil {
		d.Handler()
	}
}
