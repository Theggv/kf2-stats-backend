package maps

import "time"

type MapAnalyticsRequest struct {
	ServerId int `json:"server_id"`

	From time.Time `json:"date_from"`
	To   time.Time `json:"date_to"`

	Limit int `json:"limit"`
}

type MapAnalytics struct {
	MapId int `json:"-"`

	MapName string `json:"map_name"`
	Count   int    `json:"count"`
}

type MapAnalyticsResponse struct {
	Items []*MapAnalytics `json:"items"`
}
