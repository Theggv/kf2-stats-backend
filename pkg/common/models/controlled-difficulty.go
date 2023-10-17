package models

type CDGameData struct {
	SpawnCycle    string `json:"spawn_cycle"`
	MaxMonsters   int    `json:"max_monsters"`
	WaveSizeFakes int    `json:"wave_size_fakes"`
	ZedsType      string `json:"zeds_type"`
}
