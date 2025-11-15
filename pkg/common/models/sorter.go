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

func (c SortByRequest) Transform(lookup map[string]string, defaultValue string) (string, string) {
	field := defaultValue
	if value, ok := lookup[c.Field]; ok {
		field = value
	}

	direction := "ASC"
	if c.Direction == Desc {
		direction = "DESC"
	}

	return field, direction
}
