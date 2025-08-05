package models

type SortDirection int

const (
	Asc SortDirection = iota
	Desc
)

type SortByRequest struct {
	Field     string        `json:"field"`
	Direction SortDirection `json:"direction"`
}
