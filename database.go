package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
)

var database *sql.DB
const DBConfigFile = "dbSettings.json"
var DBConfig = DBConfigs{"","", "", ""}

func InitDatabase() {
	var err error

	var login = ""
	login += fmt.Sprintf("%s:%s", DBConfig.DBUsername, DBConfig.DBPassword)
	login += fmt.Sprintf("@tcp(%s)", DBConfig.DBPath)
	login += fmt.Sprintf("/%s", DBConfig.DBDatabase)

	database, err = sql.Open("mysql", login)
	HandleError(err, CustomError{}.Unxepected(err))
}

type DBConfigs struct {
	DBPath     string `json:"db_path"`
	DBDatabase string `json:"db_database"`
	DBUsername string `json:"db_username"`
	DBPassword string `json:"db_password"`
}

func (db *DBConfigs) init() {
	byteText, err := ioutil.ReadFile(DBConfigFile)
	HandleError(err, CustomError{}.Unxepected(err))

	err = json.Unmarshal(byteText, &db)
	HandleError(err, CustomError{}.Unxepected(err))
}

func (s Schema) InitTable() {

	var queryParams []interface{}
	var paramsDetails = ""

	queryParams = append(queryParams, s.Table)

	for _, param := range s.Params {

		paramsDetails += "?\t?\t?"

		queryParams = append(queryParams, param.Article)
		queryParams = append(queryParams, param.Type)

		if param.Null == "NO" {
			queryParams = append(queryParams, "NOT NULL")
		} else {
			queryParams = append(queryParams, "NULL")
		}

		if param.Default != "" {
			paramsDetails += "DEFAULT(?)"
			queryParams = append(queryParams, param.Default)
		}

	}

	smtp, err := database.Prepare(fmt.Sprintf("CREATE TABLE IF NOT EXISTS ? (%s)", paramsDetails))
	HandleError(err, CustomError{}.Unxepected(err))

	_, err = smtp.Exec(queryParams)
	HandleError(err, CustomError{}.Unxepected(err))
}

func (s Schema) INSERT () int {
	return 0
}

func (s Schema) SELECT () interface{} {
	var dummy interface{}
	return dummy
}

func (s Schema) UPDATE () bool {
	return false
}

func (s Schema) DELETE () bool {
	return false
}