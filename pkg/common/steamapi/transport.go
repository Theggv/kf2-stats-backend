package steamapi

type GetUserSummaryPlayer struct {
	SteamId    string `json:"steamid"`
	Name       string `json:"personaname"`
	ProfileUrl string `json:"profileurl"`
	Avatar     string `json:"avatar"`
}

type GetUserSummaryResponseResponse struct {
	Players []GetUserSummaryPlayer `json:"players"`
}

type GetUserSummaryResponse struct {
	Response *GetUserSummaryResponseResponse `json:"response"`
}
