class StatsServiceBase extends Info
	abstract;

struct ZedCounter {
	var int Cyst;
	var int AlphaClot;
	var int Slasher;
	var int Stalker;
	var int Crawler;
	var int Gorefast;
	var int Rioter;
	var int EliteCrawler;
	var int Gorefiend;

	var int Siren;
	var int Bloat;
	var int Edar;
	var int Husk;

	var int Scrake;
	var int FP;
	var int QP;
	var int Boss;

	structdefaultproperties {
		Cyst = 0
		AlphaClot = 0
		Slasher = 0
		Stalker = 0
		Crawler = 0
		Gorefast = 0
		Rioter = 0
		EliteCrawler = 0
		Gorefiend = 0
		Siren = 0
		Bloat = 0
		Edar = 0
		Husk = 0
		Scrake = 0
		FP = 0
		QP = 0
		Boss = 0
	}
};

enum AuthType {
	AT_NONE,
	AT_STEAM,
	AT_EGS
};

struct PlayerStats {
	var string PlayerName;
	var string Uid;
	var string AuthId;
	var AuthType AuthType;

	var int Perk;
	var int Level;
	var int Prestige;

	var bool IsDead;

	var int ShotsFired;
	var int ShotsHit;
	var int ShotsHS;

	var ZedCounter Kills;

	var int HuskBackpackKills;
	var int HuskRages;

	var ZedCounter InjuredBy;

	var int DoshEarned;
	
	var int HealsGiven;
	var int HealsReceived;

	var int DamageDealt;
	var int DamageTaken;

	var int ZedTimeCount;
	var float ZedTimeLength;

	structdefaultproperties {
		PlayerName = ""
		Uid = ""
		AuthId = ""
		AuthType = 0
		Perk = 0
		Level = 0
		Prestige = 0
		IsDead = false
		ShotsFired = 0
		ShotsHit = 0
		ShotsHS = 0
		DoshEarned = 0
		HealsGiven = 0
		HealsReceived = 0
		DamageDealt = 0
		DamageTaken = 0
		HuskBackpackKills = 0
		HuskRages = 0
		ZedTimeCount = 0
		ZedTimeLength = 0.0
	}
};

struct CDStruct {
	var string SpawnCycle;
	var int MaxMonsters;
	var int WaveSizeFakes;
	var string ZedsType;
};


struct CreateWaveStatsBody {
	var int SessionId; 
	var int Wave; 
	var int Length;

	var bool HasCDData;
	var CDStruct CDData;

	var array<PlayerStats> Players;

	structdefaultproperties {
		SessionId = 0
		Wave = 0
		Length = 0

		HasCDData = false
	}
};

static private function JsonObject ConvertZedStatsToJson(ZedCounter Zeds, int HuskBP) {
	local JsonObject Json;

	Json = new Class'JsonObject';
	Json.SetIntValue("cyst", Zeds.Cyst);
	Json.SetIntValue("alpha_clot", Zeds.AlphaClot);
	Json.SetIntValue("slasher", Zeds.Slasher);
	Json.SetIntValue("stalker", Zeds.Stalker);
	Json.SetIntValue("crawler", Zeds.Crawler);
	Json.SetIntValue("gorefast", Zeds.Gorefast);
	Json.SetIntValue("rioter", Zeds.Rioter);
	Json.SetIntValue("elite_crawler", Zeds.EliteCrawler);
	Json.SetIntValue("gorefiend", Zeds.Gorefiend);
	Json.SetIntValue("siren", Zeds.Siren);
	Json.SetIntValue("bloat", Zeds.Bloat);
	Json.SetIntValue("edar", Zeds.Edar);
	Json.SetIntValue("husk", Zeds.Husk - HuskBP);
	Json.SetIntValue("scrake", Zeds.Scrake);
	Json.SetIntValue("fp", Zeds.FP);
	Json.SetIntValue("qp", Zeds.QP);
	Json.SetIntValue("boss", Zeds.Boss);

	return Json;
}

static function string PrepareWaveStatsBody(CreateWaveStatsBody Body) {
	local PlayerStats PlayerData;
	local JsonObject Json, JsonPlayers, JsonPlayerData, JsonCDData;

	JsonPlayers = new Class'JsonObject';

	foreach Body.Players(PlayerData) {
		JsonPlayerData = new Class'JsonObject';
		JsonPlayerData.SetStringValue("user_name", PlayerData.PlayerName);
		JsonPlayerData.SetStringValue("user_auth_id", PlayerData.AuthId);
		JsonPlayerData.SetIntValue("user_auth_type", PlayerData.AuthType);

		JsonPlayerData.SetIntValue("perk", PlayerData.Perk);
		JsonPlayerData.SetIntValue("level", PlayerData.Level);
		JsonPlayerData.SetIntValue("prestige", PlayerData.Prestige);
		JsonPlayerData.SetBoolValue("is_dead", PlayerData.IsDead);
		JsonPlayerData.SetIntValue("shots_fired", PlayerData.ShotsFired);
		JsonPlayerData.SetIntValue("shots_hit", PlayerData.ShotsHit);
		JsonPlayerData.SetIntValue("shots_hs", PlayerData.ShotsHS);
		JsonPlayerData.SetIntValue("husk_b", PlayerData.HuskBackpackKills);
		JsonPlayerData.SetIntValue("husk_r", PlayerData.HuskRages);
		JsonPlayerData.SetIntValue("dosh_earned", PlayerData.DoshEarned);
		JsonPlayerData.SetIntValue("heals_given", PlayerData.HealsGiven);
		JsonPlayerData.SetIntValue("heals_recv", PlayerData.HealsReceived);
		JsonPlayerData.SetIntValue("damage_dealt", PlayerData.DamageDealt);
		JsonPlayerData.SetIntValue("damage_taken", PlayerData.DamageTaken);
		JsonPlayerData.SetIntValue("zedtime_count", PlayerData.ZedTimeCount);
		JsonPlayerData.SetFloatValue("zedtime_length", PlayerData.ZedTimeLength);

		JsonPlayerData.SetObject("kills", ConvertZedStatsToJson(PlayerData.Kills, PlayerData.HuskBackpackKills));
		JsonPlayerData.SetObject("injured_by", ConvertZedStatsToJson(PlayerData.InjuredBy, 0));

		JsonPlayers.ObjectArray.AddItem(JsonPlayerData);
	}

	Json = new Class'JsonObject';
	Json.SetIntValue("session_id", Body.SessionId);
	Json.SetIntValue("wave", Body.Wave);
	Json.SetIntValue("wave_length", Body.Length);

	if (Body.HasCDData) {
		JsonCDData = new Class'JsonObject';
		JsonCDData.SetStringValue("spawn_cycle", Body.CDData.SpawnCycle);
		JsonCDData.SetIntValue("max_monsters", Body.CDData.MaxMonsters);
		JsonCDData.SetIntValue("wave_size_fakes", Body.CDData.WaveSizeFakes);
		JsonCDData.SetStringValue("zeds_type", Body.CDData.ZedsType);
		Json.SetObject("cd_data", JsonCDData);
	}

	Json.SetObject("players", JsonPlayers);
	
	return Class'JsonObject'.static.EncodeJson(Json);
}

function CreateWaveStats(CreateWaveStatsBody Body);

