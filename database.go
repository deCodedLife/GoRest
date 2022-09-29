package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	_ "github.com/go-sql-driver/mysql"
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

	var query = ""

	for _, param := range s.Params {

		query += fmt.Sprintf("`%s` ", param.Article)
		query += param.Type + " "

		if param.Null == "NO" {
			query += "NOT NULL"
		} else {
			query += "NULL"
		}

		if param.Default != "" {
			query += " "
			paramPart1 := strings.Split(param.Type, "bit")
			paramPart2 := strings.Split(param.Type, "int")

			if len(paramPart1) > 1 || len(paramPart2) > 1 {
				query += fmt.Sprintf("DEFAULT %s", param.Default)
			} else {
				query += fmt.Sprintf("DEFAULT '%s'", param.Default)
			}
		}

		query += ", "
	}

	query = query[:len(query) - 2]
	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s)", s.Table, query)

	smtp, err := database.Prepare(query)
	HandleError(err, CustomError{}.Unxepected(err))

	_, err = smtp.Exec()
	HandleError(err, CustomError{}.Unxepected(err))
}

func (s Schema) INSERT () int64 {
	var columns = ""
	var values = ""
	var queryParams []interface{}

	for _, param := range s.Params {

		columns = "?,"
		values = "?,"

		queryParams = append(queryParams, param.Article)

	}

	columns = columns[:len(columns) - 1]
	values = values[:len(values) - 1]

	stmp, err := database.Prepare(fmt.Sprintf("INSERT INTO ? (%s) VALUES", columns, values))
	HandleError(err, CustomError{}.Unxepected(err))

	data, err := stmp.Exec(queryParams)
	HandleError(err, CustomError{}.Unxepected(err))

	insertedID, err := data.LastInsertId()
	return insertedID
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