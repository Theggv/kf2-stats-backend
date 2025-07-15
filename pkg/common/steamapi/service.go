package steamapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chenyahui/gin-cache/persist"
)

const (
	baseUrl = "https://api.steampowered.com"
)

type SteamApiUserService struct {
	apiKey string
	client *http.Client

	memoryStore *persist.MemoryStore
}

type SteamUser struct {
	AuthId     string
	ProfileUrl *string
	Avatar     *string
}

func NewSteamApiUserService(apiKey string) *SteamApiUserService {
	return &SteamApiUserService{
		apiKey:      apiKey,
		client:      &http.Client{},
		memoryStore: persist.NewMemoryStore(5 * time.Minute),
	}
}

func (s *SteamApiUserService) GetUserSummary(steamIds []string) ([]GetUserSummaryPlayer, error) {
	summaries := []GetUserSummaryPlayer{}
	chunkSize := 100

	var cached GetUserSummaryPlayer
	chunk := []string{}
	for _, steamId := range steamIds {
		if err := s.memoryStore.Get(steamId, &cached); err == nil {
			summaries = append(summaries, cached)
			continue
		}

		chunk = append(chunk, steamId)

		if len(chunk) == chunkSize {
			data, err := s.getUsersSummaryInternal(chunk)
			if err != nil {
				fmt.Printf("warn: %v\n", err)
				return summaries, nil
			}

			summaries = append(summaries, data...)
			chunk = chunk[:0]
		}
	}

	if len(chunk) > 0 {
		data, err := s.getUsersSummaryInternal(chunk)
		if err != nil {
			fmt.Printf("warn: %v\n", err)
			return summaries, nil
		}

		summaries = append(summaries, data...)
	}

	for _, item := range summaries {
		s.memoryStore.Set(item.SteamId, item, 5*time.Minute)
	}

	return summaries, nil
}

func (s *SteamApiUserService) getUsersSummaryInternal(steamIds []string) ([]GetUserSummaryPlayer, error) {
	url := fmt.Sprintf("%v/ISteamUser/GetPlayerSummaries/v0002/?key=%v&steamids=%v",
		baseUrl, s.apiKey, strings.Join(steamIds, ","),
	)

	res, err := s.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("getUsersSummaryInternal: %v", res.Status)
	}

	var resJson GetUserSummaryResponse
	if err := json.NewDecoder(res.Body).Decode(&resJson); err != nil {
		return nil, err
	}

	if resJson.Response == nil {
		return nil, errors.New("getUsersSummaryInternal: no response")
	}

	return resJson.Response.Players, nil
}

func (s *SteamApiUserService) ValidateOpenId(req ValidateOpenIdRequest) (*GetUserSummaryPlayer, error) {
	req.Params = strings.ReplaceAll(req.Params, "id_res", "check_authentication")

	endpoint := "https://steamcommunity.com/openid/login"
	data := []byte(req.Params)

	r, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	r.Header.Add("Accept-Language", "en")
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", fmt.Sprintf("%v", len(data)))

	res, err := s.client.Do(r)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	responseText := string(bytes)
	if !strings.Contains(responseText, "is_valid:true") {
		return nil, errors.New("is_valid:false")
	}

	steamId, ok := tryGetSteamId(req.Params)
	if !ok {
		return nil, errors.New("invalid steamid")
	}

	summary, err := s.GetUserSummary([]string{steamId})
	if err != nil {
		return nil, err
	}

	return &summary[0], nil
}
