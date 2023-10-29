package util

import "github.com/theggv/kf2-stats-backend/pkg/common/models"

func ParsePagination(pager models.PaginationRequest) (int, int) {
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
