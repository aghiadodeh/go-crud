package models

type ListResponse[T any] struct {
	Total    int64 `json:"total"`
	Data     []T   `json:"data"`
	Metadata any   `json:"metadata,omitempty"`
}
