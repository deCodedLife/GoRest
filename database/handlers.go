package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	. "backend/tool"
)

var database *sql.DB
var DBConfig = DBConfigs{"", "", "", ""}

const DBConfigFile = "dbSettings.json"

func InitDatabase() {
	var err error

	var login = ""
	login += fmt.Sprintf("%s:%s", DBConfig.DBUsername, DBConfig.DBPassword)
	login += fmt.Sprintf("@tcp(%s)", DBConfig.DBPath)
	login += fmt.Sprintf("/%s", DBConfig.DBDatabase)

	database, err = sql.Open("mysql", login)
	HandleError(err, CustomError{}.Unxepected(err))
}

func (db *DBConfigs) Init() {
	byteText, err := ioutil.ReadFile(DBConfigFile)
	HandleError(err, CustomError{}.Unxepected(err))

	err = json.Unmarshal(byteText, &db)
	HandleError(err, CustomError{}.Unxepected(err))
}

func (s Schema) InitTable() {

	var additional = ""
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

			var numericParams []int
			isNumeric := false

			numericParams = append(numericParams, len(strings.Split(param.Type, "bit")))
			numericParams = append(numericParams, len(strings.Split(param.Type, "int")))
			numericParams = append(numericParams, len(strings.Split(param.Type, "double")))

			for _, value := range numericParams {
				if value > 1 {
					isNumeric = true
				}
			}

			if isNumeric {
				query += fmt.Sprintf("DEFAULT %s", param.Default)
			} else {
				query += fmt.Sprintf("DEFAULT '%s'", param.Default)
			}
		}

		if strings.ToLower(param.Article) == "id" {
			query += " UNIQUE AUTO_INCREMENT"
			additional = fmt.Sprintf(", PRIMARY KEY (%s)", param.Article)
		}

		query += ", "
	}

	query = query[:len(query)-2]
	query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (%s%s)", s.Table, query, additional)

	stmt, err := database.Prepare(query)
	HandleError(err, CustomError{}.Unxepected(err))

	_, err = stmt.Exec()
	HandleError(err, CustomError{}.Unxepected(err))
}

func (s Schema) INSERT(d map[string]interface{}) (int64, error) {
	var columns = ""
	var values = ""
	var queryParams []interface{}

	for _, param := range s.Params {

		if d[param.Article] == nil && param.Null != "NO" {
			continue
		}

		if strings.ToLower(param.Article) == "id" {
			continue
		}

		columns += fmt.Sprintf("`%s`, ", param.Article)

		if d[param.Article] == nil {

			if param.IsNumeric() {
				values += fmt.Sprintf("%v, ", param.Default)
			} else {
				values += fmt.Sprintf("'%v', ", param.Default)
			}

			continue
		}

		if param.IsNumeric() {
			//values += "?, "
			values += fmt.Sprintf("%v, ", d[param.Article])
		} else {
			values += fmt.Sprintf("'%v', ", d[param.Article])
			//values += "'?', "

		}

		queryParams = append(queryParams, fmt.Sprintf("%v", d[param.Article]))
	}

	columns = columns[:len(columns)-2]
	values = values[:len(values)-2]

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", s.Table, columns, values)
	stmt, err := database.Prepare(query)
	if err != nil {
		return 0, err
	}

	data, err := stmt.Exec()
	if err != nil {
		return 0, err
	}

	insertedID, err := data.LastInsertId()
	return insertedID, nil
}

func (s Schema) SELECT() interface{} {
	var dummy interface{}
	return dummy
}

func (s Schema) UPDATE() bool {
	return false
}

func (s Schema) DELETE() bool {
	return false
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

func (s Schema) ValidateParams(d map[string]interface{}) error {
	for _, param := range s.Params {

		if param.Null != "NO" || strings.ToLower(param.Article) == "id" {
			continue
		}

		if param.Default != "" {
			continue
		}

		if d[param.Article] == nil {
			return errors.New(fmt.Sprintf("%s is required", param.Article))
		}

	}

	return nil
}
