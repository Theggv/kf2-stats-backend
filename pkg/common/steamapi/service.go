package steamapi

import (
	"encoding/json"
	"errors"
	"fmt"
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
		chunk = chunk[:0]
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

	var resJson GetUserSummaryResponse
	if err := json.NewDecoder(res.Body).Decode(&resJson); err != nil {
		return nil, err
	}

	if resJson.Response == nil {
		return nil, errors.New("response is null")
	}

	return resJson.Response.Players, nil
}
