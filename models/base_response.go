package models

type BaseResponse[T any] struct {
	Success    bool   `json:"success"`
	Data       T      `json:"data"`
	Message    string `json:"message"`
	StatusCode int    `json:"statusCode"`
	Error      string `json:"error,omitempty"`
}
