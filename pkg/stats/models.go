package stats

type ZedCounter struct {
	Cyst         int `json:"cyst"`
	AlphaClot    int `json:"alpha_clot"`
	Slasher      int `json:"slasher"`
	Stalker      int `json:"stalker"`
	Crawler      int `json:"crawler"`
	Gorefast     int `json:"gorefast"`
	Rioter       int `json:"rioter"`
	EliteCrawler int `json:"elite_crawler"`
	Gorefiend    int `json:"gorefiend"`

	Siren int `json:"siren"`
	Bloat int `json:"bloat"`
	Edar  int `json:"edar"`
	Husk  int `json:"husk"`

	Scrake int `json:"scrake"`
	FP     int `json:"fp"`
	QP     int `json:"qp"`
	Boss   int `json:"boss"`
	Custom int `json:"custom"`
}
