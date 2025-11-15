package models

type PaginationRequest struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"results_per_page"`
}

func (pager PaginationRequest) Parse() (int, int) {
	page := pager.Page
	resultsPerPage := pager.ResultsPerPage

	if page < 0 {
		page = 0
	}

	if resultsPerPage < 10 {
		resultsPerPage = 10
	}

	if resultsPerPage > 100 {
		resultsPerPage = 100
	}

	return page, resultsPerPage
}

type PaginationResponse struct {
	Page           int `json:"page"`
	ResultsPerPage int `json:"results_per_page"`
	TotalResults   int `json:"total_results"`
}
