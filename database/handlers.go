package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	. "github.com/deCodedLife/gorest/tool"
)

var database *sql.DB
var DBConfig = DBConfigs{"", "", "", ""}

func InitDatabase() {
	var err error

	var login = ""
	login += fmt.Sprintf("%s:%s", DBConfig.DBUsername, DBConfig.DBPassword)
	login += fmt.Sprintf("@tcp(%s)", DBConfig.DBPath)
	login += fmt.Sprintf("/%s", DBConfig.DBDatabase)

	database, err = sql.Open("mysql", login)
	HandleError(err, CustomError{}.Unexpected(err))
}

func (db *DBConfigs) Init() {
	byteText, err := os.ReadFile(DBConfigFile)
	HandleError(err, CustomError{}.Unexpected(err))

	err = json.Unmarshal(byteText, &db)
	HandleError(err, CustomError{}.Unexpected(err))
}

func (s *Schema) InitTable() {
	var additional = ""
	var query = ""
	(*s).ParamsCount = 0

	for _, param := range s.Params {
		value := reflect.ValueOf(param)
		field := value.FieldByName("Type")

		if field.IsValid() == false {
			continue
		}

		(*s).ParamsCount++
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
	HandleError(err, CustomError{}.Unexpected(err))

	_, err = stmt.Exec()
	HandleError(err, CustomError{}.Unexpected(err))
}

func (s Schema) INSERT(d map[string]interface{}) (int64, error) {
	var columns = ""
	var values = ""

	for _, param := range s.Params {
		value := reflect.ValueOf(param)
		field := value.FieldByName("Type")

		if field.IsValid() == false {
			continue
		}

		if d[param.Article] == nil {

			if param.Null != "NO" {
				continue
			}

			if param.Default != "" {
				continue
			}

		}

		if strings.ToLower(param.Article) == "id" {
			continue
		}

		columns += fmt.Sprintf("`%s`, ", param.Article)

		if param.IsNumeric() {
			values += fmt.Sprintf("%v, ", d[param.Article])
			continue
		}

		values += fmt.Sprintf("'%v', ", d[param.Article])

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

func (s Schema) SELECT(d map[string]interface{}) ([]map[string]interface{}, error) {
	var response []map[string]interface{}
	var params = make([]SchemaParam, s.ParamsCount)

	var responsePointers = make([]interface{}, s.ParamsCount)
	var responseColumns = make([]interface{}, s.ParamsCount)
	var whereClauses = "WHERE "

	for index, param := range s.Params {
		value := reflect.ValueOf(param)
		field := value.FieldByName("Type")

		if field.IsValid() == false {
			continue
		}

		params = append(params, param)
		responseColumns[index] = &responsePointers[index]

		if d[param.Article] != nil {

			if param.IsNumeric() {
				whereClauses += fmt.Sprintf("`%s` = %v AND ", param.Article, d[param.Article])
				continue
			}

			whereClauses += fmt.Sprintf("`%s` LIKE '%%%v%%' AND ", param.Article, d[param.Article])

		}

	}

	if whereClauses == "WHERE " {
		whereClauses = ""
	} else {
		whereClauses = whereClauses[:len(whereClauses)-len(" AND ")]
	}

	query := fmt.Sprintf("SELECT * FROM %s %s", s.Table, whereClauses)
	stmt, err := database.Prepare(query)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(responseColumns...)

		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})

		log.Println(params)
		log.Println(responsePointers)

		for i, value := range responsePointers {

			param := params[i]
			valueString := fmt.Sprintf("%s", value)

			if len(strings.Split(param.Type, "int")) > 1 {
				intVal, err := strconv.Atoi(valueString)
				if err != nil {
					continue
				}
				row[param.Article] = intVal
				continue
			}

			if len(strings.Split(param.Type, "float")) > 1 {
				floatVal, err := strconv.ParseFloat(valueString, 64)
				if err != nil {
					continue
				}
				row[param.Article] = floatVal
				continue
			}

			if len(strings.Split(param.Type, "char")) > 1 {
				if valueString == "Y" {
					row[param.Article] = true
				} else if valueString == "N" {
					row[param.Article] = false
				}
			}

			row[param.Article] = valueString
		}

		response = append(response, row)

		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func nextParam(s Schema, index int) int {
	if s.Params[index].Type == "" {
		return nextParam(s, index+1)
	}
	return index
}

func (s Schema) UPDATE(id int, d map[string]interface{}) (map[string]interface{}, error) {
	var setClause = ""

	for _, param := range s.Params {
		value := reflect.ValueOf(param)
		field := value.FieldByName("Type")

		if field.IsValid() == false {
			continue
		}

		if d[param.Article] == nil {
			continue
		}

		if param.IsNumeric() {
			setClause += fmt.Sprintf("`%s` = %v, ", param.Article, d[param.Article])
			continue
		}

		setClause += fmt.Sprintf("`%s` = '%s', ", param.Article, d[param.Article])

	}

	if setClause == "" {
		return nil, errors.New("not allowed")
	}

	setClause = setClause[:len(setClause)-2]

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = %d", s.Table, setClause, id)
	stmt, err := database.Prepare(query)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"rowsAffected": rowsAffected,
	}, nil
}

func (s Schema) DELETE(id int) (map[string]interface{}, error) {

	if id == 0 {
		return nil, errors.New("not allowed")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = %d", s.Table, id)
	stmt, err := database.Prepare(query)
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec()
	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"rowsAffected": rowsAffected,
	}, nil

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

func (s Schema) ContainsMethod(m string) bool {

	for _, method := range s.Methods {

		if strings.ToUpper(method) == m {
			return true
		}

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
