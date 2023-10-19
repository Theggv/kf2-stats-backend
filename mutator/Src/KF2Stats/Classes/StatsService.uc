class StatsService extends StatsServiceBase;

static function StatsService GetInstance() {
	local StatsService Instance;

	foreach Class'WorldInfo'.static.GetWorldInfo().DynamicActors(Class'StatsService', Instance) {      
		return Instance;        
	}

	return Instance;
}

public function CreateWaveStats(CreateWaveStatsBody Body) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/stats/wave",
		"POST",
		PrepareWaveStatsBody(Body)
	);

	R.ProcessRequest();
}

defaultproperties {
}