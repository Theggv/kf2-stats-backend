package models

type SessionMetadata struct {
	Difficulty *SessionMetadataDifficulty `json:"diff,omitempty"`
}

type SessionMetadataDifficultyWave struct {
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

type SessionMetadataDifficultySummary struct {
	AvgZedsDifficulty float64 `json:"avg_zeds_difficulty"`
	MapBonus          float64 `json:"map_bonus"`

	CompletionPercent float64 `json:"completion_p"`
	RestartsPenalty   float64 `json:"restarts_penalty"`

	PotentialScore float64 `json:"potential_score"`
	FinalScore     float64 `json:"final_score"`
}

type SessionMetadataDifficulty struct {
	SessionId int `json:"session_id"`

	Summary *SessionMetadataDifficultySummary `json:"summary"`
	Waves   []*SessionMetadataDifficultyWave  `json:"waves"`
}
