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

func (c ZedCounter) GetTotalMediums() int {
	return c.Rioter + c.Gorefiend + c.Scrake + c.Bloat + c.Edar + c.Husk
}

func (c ZedCounter) GetTotalLarges() int {
	return c.Scrake + c.FP + c.QP
}

func (c ZedCounter) ConvertToMap() map[string]int {
	data := map[string]int{
		"cyst":          c.Cyst,
		"alpha_clot":    c.AlphaClot,
		"slasher":       c.Slasher,
		"stalker":       c.Stalker,
		"crawler":       c.Crawler,
		"gorefast":      c.Gorefast,
		"elite_crawler": c.EliteCrawler,

		"rioter":    c.Rioter,
		"gorefiend": c.Gorefiend,
		"edar":      c.Edar,

		"siren": c.Siren,
		"bloat": c.Bloat,
		"husk":  c.Husk,

		"scrake": c.Scrake,
		"fp":     c.FP,
		"qp":     c.QP,

		"boss":   c.Boss,
		"custom": c.Custom,
	}

	return data
}
