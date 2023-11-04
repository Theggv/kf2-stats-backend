class StatsServiceBase extends Info
	abstract;

struct RadioCounter {
	var int RequestHealing;
	var int RequestDosh;
	var int RequestHelp;
	var int TauntZeds;
	var int FollowMe;
	var int GetToTheTrader;
	var int Affirmative;
	var int Negative;
	var int ThankYou;

	structdefaultproperties {
		RequestHealing = 0
		RequestDosh = 0
		RequestHelp = 0
		TauntZeds = 0
		FollowMe = 0
		GetToTheTrader = 0
		Affirmative = 0
		Negative = 0
		ThankYou = 0
	}
};

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

	var int Custom;

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
		Custom = 0
	}
};

enum AuthType {
	AT_NONE,
	AT_STEAM,
	AT_EGS
};


struct PlayerDataStats {
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

	var RadioCounter RadioComms;

	structdefaultproperties {
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

struct PlayerData {
	var string UniqueId;
	
	var string PlayerName;
	var string AuthId;
	var AuthType AuthType;

	var int Perk;
	var int Level;
	var int Prestige;

	var int Health;
	var int Armor;

	var bool IsDead;
	var bool IsSpectator;

	var PlayerDataStats Stats;

	structdefaultproperties {
		UniqueId = ""
		PlayerName = ""
		AuthId = ""
		AuthType = 0

		Perk = 0
		Level = 0
		Prestige = 0

		Health = 0
		Armor = 0

		IsDead = false
		IsSpectator = false
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

	var array<PlayerData> Players;

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
	Json.SetIntValue("custom", Zeds.Custom);

	return Json;
}

static function string PrepareWaveStatsBody(CreateWaveStatsBody Body) {
	local PlayerData D;
	local JsonObject Json, JsonPlayers, JsonPlayerData, JsonCDData;

	JsonPlayers = new Class'JsonObject';

	foreach Body.Players(D) {
		JsonPlayerData = new Class'JsonObject';
		JsonPlayerData.SetStringValue("user_name", D.PlayerName);
		JsonPlayerData.SetStringValue("user_auth_id", D.AuthId);
		JsonPlayerData.SetIntValue("user_auth_type", D.AuthType);

		JsonPlayerData.SetIntValue("perk", D.Perk);
		JsonPlayerData.SetIntValue("level", D.Level);
		JsonPlayerData.SetIntValue("prestige", D.Prestige);
		JsonPlayerData.SetBoolValue("is_dead", D.IsDead);
		JsonPlayerData.SetIntValue("shots_fired", D.Stats.ShotsFired);
		JsonPlayerData.SetIntValue("shots_hit", D.Stats.ShotsHit);
		JsonPlayerData.SetIntValue("shots_hs", D.Stats.ShotsHS);
		JsonPlayerData.SetIntValue("husk_b", D.Stats.HuskBackpackKills);
		JsonPlayerData.SetIntValue("husk_r", D.Stats.HuskRages);
		JsonPlayerData.SetIntValue("dosh_earned", D.Stats.DoshEarned);
		JsonPlayerData.SetIntValue("heals_given", D.Stats.HealsGiven);
		JsonPlayerData.SetIntValue("heals_recv", D.Stats.HealsReceived);
		JsonPlayerData.SetIntValue("damage_dealt", D.Stats.DamageDealt);
		JsonPlayerData.SetIntValue("damage_taken", D.Stats.DamageTaken);
		JsonPlayerData.SetIntValue("zedtime_count", D.Stats.ZedTimeCount);
		JsonPlayerData.SetFloatValue("zedtime_length", D.Stats.ZedTimeLength);

		JsonPlayerData.SetIntValue("request_healing", D.Stats.RadioComms.RequestHealing);
		JsonPlayerData.SetIntValue("request_dosh", D.Stats.RadioComms.RequestDosh);
		JsonPlayerData.SetIntValue("request_help", D.Stats.RadioComms.RequestHelp);
		JsonPlayerData.SetIntValue("taunt_zeds", D.Stats.RadioComms.TauntZeds);
		JsonPlayerData.SetIntValue("follow_me", D.Stats.RadioComms.FollowMe);
		JsonPlayerData.SetIntValue("get_to_the_trader", D.Stats.RadioComms.GetToTheTrader);
		JsonPlayerData.SetIntValue("affirmative", D.Stats.RadioComms.Affirmative);
		JsonPlayerData.SetIntValue("negative", D.Stats.RadioComms.Negative);
		JsonPlayerData.SetIntValue("thank_you", D.Stats.RadioComms.ThankYou);

		JsonPlayerData.SetObject("kills", ConvertZedStatsToJson(D.Stats.Kills, D.Stats.HuskBackpackKills));
		JsonPlayerData.SetObject("injured_by", ConvertZedStatsToJson(D.Stats.InjuredBy, 0));

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

