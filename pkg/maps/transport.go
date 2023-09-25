package maps

type AddMapRequest struct {
	Name    string `json:"name" binding:"required"`
	Preview string `json:"preview"`
}

type AddMapResponse struct {
	Id int `json:"id"`
}

type GetByPatternResponse struct {
	Items []Map `json:"items"`
}

type UpdatePreviewRequest struct {
	Id      int    `json:"id"`
	Preview string `json:"preview"`
}
