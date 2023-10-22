package steamapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	baseUrl = "https://api.steampowered.com"
)

type SteamApiUserService struct {
	apiKey string
	client *http.Client
}

func NewSteamApiUserService(apiKey string) *SteamApiUserService {
	return &SteamApiUserService{
		apiKey: apiKey,
		client: &http.Client{},
	}
}

func (s *SteamApiUserService) GetUserSummary(steamIds []string) ([]GetUserSummaryPlayer, error) {
	url := fmt.Sprintf("%v/ISteamUser/GetPlayerSummaries/v0002/?key=%v&steamids=%v",
		baseUrl, s.apiKey, strings.Join(steamIds, ","),
	)

	res, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var resJson GetUserSummaryResponse
	if err := json.NewDecoder(res.Body).Decode(&resJson); err != nil {
		return nil, err
	}

	if resJson.Response == nil {
		return nil, errors.New("response is null")
	}

	return resJson.Response.Players, nil
}
