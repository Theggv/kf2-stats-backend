package models

type PaginationRequest struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"results_per_page"`
}

type PaginationResponse struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"results_per_page"`
	TotalResults   int `json:"total_results"`
}
