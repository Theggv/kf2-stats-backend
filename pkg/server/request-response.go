package server

type AddServerRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
	Type    any    `json:"type" binding:"required,numeric,gte=0"`
}

type AddServerResponse struct {
	Id int `json:"id"`
}

type GetByPatternResponse struct {
	Items []Server `json:"items"`
}
