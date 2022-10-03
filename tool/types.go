package tool

type CustomError struct {
	StatusCode int32 `json:"status_code"`
	Handler    func()
}

type Response struct {
	StatusCode int32       `json:"status_code"`
	Data       interface{} `json:"data"`
}
