package database

type DBConfigs struct {
	DBPath     string `json:"db_path"`
	DBDatabase string `json:"db_database"`
	DBUsername string `json:"db_username"`
	DBPassword string `json:"db_password"`
}

type SchemaParam struct {
	Title       string `json:"title"`
	Article     string `json:"article"`
	Type        string `json:"type"`
	Null        string `json:"null"`
	Default     string `json:"default"`
	Display     bool   `json:"display"`
	DisplayType string `json:"display_type"`
	TakeFrom    string `json:"take_from"`
	Join        string `json:"join"`
}

type Schema struct {
	Title       string        `json:"title"`
	Table       string        `json:"table"`
	Methods     []string      `json:"methods"`
	Params      []SchemaParam `json:"params"`
	ParamsCount int           `json:"paramsCount"`
}
