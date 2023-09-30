package server

type AddServerRequest struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
}

type AddServerResponse struct {
	Id int `json:"id"`
}

type GetByPatternResponse struct {
	Items []Server `json:"items"`
}

type UpdateNameRequest struct {
	Id   int    `json:"id"`
	Name string `json:"name" binding:"required"`
}
