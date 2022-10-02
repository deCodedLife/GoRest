***unstable***

## GO REST
It's a powerful tool for deploying your database for REST api.
All what you need is a schema of your database!

This project automatically creates tables from raw json files and handles GET, POST, 
PUT, DELETE methods for them 

---

## Build from scratch
If you don't have a go language installed first visit https://go.dev/dl/ 
and setup golang first
---
* Download repo
```shell script
git clone https://github.com/deCodedLife/GoRest
```
* Get all dependencies for this project and build project
```shell script
go get -u -v -f all
go build
```
* Create configuration file `dbSettings.json` for your database and write your authorisation data
```json5
{
  "db_path": "localhost:3306",    // Address of your database
  "db_database": "test",          // Database name
  "db_username": "root",          // Database username
  "db_password": "admin"          // Database password
}
```
* Create a folder "schema"
```shell script
mkdir schema
```
* Create files which describe a schema of your database
```json5
{
  "title": "Test table",          // Table description
  "table": "test_table",          // Table name
  "params": [                     // List of columns
    {
      "title": "Identification",  // Column description
      "article": "id",            // Column name
      "type": "int",              // Column type (same as database types)
      "null": "NO",               // Can be null
      "default": ""               // Default column value
    },
    ...
  ]
}
```
*You can set max length of column value using type like `int(11)`. You should put only `YES` or `NO` in a `null
` parameter* 
* Launch your application!

---

### Todo
- [X] Handle POST method
- [ ] Handle GET method
- [ ] Handle PUT method
- [ ] Handle DELETE methods
- [ ] Clean up code
