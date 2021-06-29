package errors

import "encoding/json"

type JsonRPCError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func (err *JsonRPCError) Error() string {
	e, _ := json.Marshal(err)
	return string(e)
}

func New(code int, msg string) *JsonRPCError  {
	return &JsonRPCError{
		Code: code,
		Message: msg,
	}
}