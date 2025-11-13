package difficulty

import (
	"github.com/theggv/kf2-stats-backend/pkg/common/models"
)

type DifficultyCalculatorGameScore struct {
	AvgZedsDifficulty    float64 `json:"avg_zeds_difficulty"`
	StdDevZedsDifficulty float64 `json:"stddev_zeds_difficulty"`
	MinZedsDifficulty    float64 `json:"min_zeds_difficulty"`
	MaxZedsDifficulty    float64 `json:"max_zeds_difficulty"`

	CompletionPercent float64 `json:"completion_p"`
	RestartsPenalty   float64 `json:"restarts_penalty"`

	PotentialScore float64 `json:"potential_score"`
	FinalScore     float64 `json:"final_score"`
}

type DifficultyCalculatorGameSession struct {
	Id       int `json:"id"`
	ServerId int `json:"server_id"`
	MapId    int `json:"map_id"`

	Mode       models.GameMode       `json:"mode"`
	Status     models.GameStatus     `json:"status"`
	Length     models.GameLength     `json:"length"`
	Difficulty models.GameDifficulty `json:"diff"`
}

type DifficultyCalculatorGameWaveScore struct {
	ZedsDifficulty float64 `json:"zeds_difficulty"`

	PredictedDuration      float64 `json:"predicted_duration"`
	PredictedDurationError float64 `json:"predicted_duration_err"`

	KitingPenalty       float64 `json:"kiting_penalty"`
	WaveSizePenalty     float64 `json:"wave_size_penalty"`
	TotalPlayersPenalty float64 `json:"total_players_penalty"`

	Score float64 `json:"score"`
}

type DifficultyCalculatorGameWave struct {
	Id int `json:"id"`

	Wave    int `json:"wave"`
	Attempt int `json:"attempt"`

	Duration         int     `json:"duration"`
	DurationRealtime float64 `json:"duration_realtime"`

	TotalPlayers int `json:"total_players"`
	TotalDeaths  int `json:"total_deaths"`

	ZedtimeLength float64 `json:"zedtime_length"`
	ZedtimeCount  int     `json:"zedtime_count"`

	MaxMonsters int    `json:"max_monsters"`
	SpawnCycle  string `json:"spawn_cycle"`
	ZedsType    string `json:"zeds_type"`

	Zeds *models.ZedCounter `json:"-"`

	TotalZeds     int     `json:"total_zeds"`
	MediumPercent float64 `json:"medium_p"`
	LargePercent  float64 `json:"large_p"`

	Result *DifficultyCalculatorGameWaveScore `json:"result"`
}

type DifficultyCalculatorGame struct {
	Session *DifficultyCalculatorGameSession `json:"session"`
	Waves   []*DifficultyCalculatorGameWave  `json:"waves"`
	Result  *DifficultyCalculatorGameScore   `json:"result"`
}

func (d *DifficultyCalculatorGame) GetLastWave() (*DifficultyCalculatorGameWave, bool) {
	if len(d.Waves) > 0 {
		return d.Waves[len(d.Waves)-1], true
	}

	return nil, false
}

type GetSessionDifficultyResponseWave struct {
	WaveId int `json:"wave_id"`

	ZedsDifficulty float64 `json:"zeds_difficulty"`

	Duration               float64 `json:"duration"`
	PredictedDuration      float64 `json:"predicted_duration"`
	PredictedDurationError float64 `json:"predicted_duration_err"`

	KitingPenalty       float64 `json:"kiting_penalty"`
	WaveSizePenalty     float64 `json:"wave_size_penalty"`
	TotalPlayersPenalty float64 `json:"total_players_penalty"`

	Score float64 `json:"score"`
}

type GetSessionDifficultyResponseSummary struct {
	AvgZedsDifficulty float64 `json:"avg_zeds_difficulty"`
	MapBonus          float64 `json:"map_bonus"`

	CompletionPercent float64 `json:"completion_p"`
	RestartsPenalty   float64 `json:"restarts_penalty"`

	PotentialScore float64 `json:"potential_score"`
	FinalScore     float64 `json:"final_score"`
}

type GetSessionDifficultyResponse struct {
	SessionId int `json:"session_id"`

	Summary *GetSessionDifficultyResponseSummary `json:"summary"`
	Waves   []*GetSessionDifficultyResponseWave  `json:"waves"`
}
