class StatsRepo extends Info
	dependson (StatsServiceBase);

enum EventType {
	ET_HUSK_BACKPACK,
	ET_RAGED_BY_BP
};

enum SessionStatus {
	SESSION_STATUS_LOBBY,
	SESSION_STATUS_INPROGRESS,
	SESSION_STATUS_WON,
	SESSION_STATUS_LOST,
};

struct SessionStruct {
	var int GameMode;
	var int GameLength;
	var int GameDifficulty;
	var SessionStatus GameStatus;

	var int ServerId;
	var int MapId;
	var int SessionId;

	var int Wave;
	var bool IsActive;
	var float WaveStartedAt;

	var float ZedTimeDuration;
	var int ZedTimeCount;

	structdefaultproperties {
		GameMode = 0
		GameLength = 0
		GameDifficulty = 0
		GameStatus = 0

		ServerId = 0
		MapId = 0
		SessionId = 0

		Wave = 1
		IsActive = false
		WaveStartedAt = 0
		ZedTimeDuration = 0
		ZedTimeCount = 0
	}
};

var array<PlayerStats> Players;
var SessionStruct SessionData;

var private WorldInfo WI;
var private KFGameInfo KFGI;
var private KFGameReplicationInfo KFGRI;
var private OnlineSubsystem OS;

function PostBeginPlay() {
	super.PostBeginPlay();

	if (Role != ROLE_Authority) return;

	PostInit();
}

private function PostInit() {
	WI = Class'WorldInfo'.static.GetWorldInfo();
	if (WI == None) {
		SetTimer(0.1, false, 'PostInit');
		return;
	}

	KFGI = KFGameInfo(WI.Game);
	if (KFGI == None) {
		SetTimer(0.1, false, 'PostInit');
		return;
	}

	KFGRI = KFGI.MyKFGRI;
	if (KFGRI == None) {
		SetTimer(0.1, false, 'PostInit');
		return;
	}

	OS = class'GameEngine'.static.GetOnlineSubsystem();
	if (OS == None) {
		SetTimer(0.1, false, 'PostInit');
		return;
	}

	SetTimer(0.1, true, 'UpdateWave');
	SetTimer(15.0, true, 'UpdateGameData');
}

private function UpdateWave() {
	local int currentWave;
	local bool isWaveActive;

	currentWave = KFGRI.WaveNum;
	isWaveActive = KFGI.IsWaveActive();

	if (isWaveActive != SessionData.IsActive) {
		if (isWaveActive) {
			if (currentWave != SessionData.Wave) {
				SessionData.Wave = currentWave;
			}

			OnWaveStarted();
		} else {
			OnWaveEnded();
		}

		SessionData.IsActive = isWaveActive;
	}
}

private function OnWaveStarted() {
    local KFPlayerController C;

	SessionData.WaveStartedAt = WI.RealTimeSeconds;
	SessionData.ZedTimeDuration = 0.0;
	SessionData.ZedTimeCount = 0;

	if (SessionData.SessionId == 0) return;

	if (SessionData.Wave == 1) {
		class'SessionService'.static.GetInstance().UpdateStatus(
			SessionData.SessionId,
			SESSION_STATUS_INPROGRESS
		);
	}

	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;
		
		ResetPlayerStats(C);
	}
}

private function OnWaveEnded() {
    local KFPlayerController C;
	local int Alive;

	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;

		UpdateNonKillStats(C);
	}

	if (SessionData.SessionId == 0) return;
	
	UploadWaveStats();

	Alive = GetAliveCount();

	if (Alive == 0) {
		class'SessionService'.static.GetInstance().UpdateStatus(
			SessionData.SessionId,
			SESSION_STATUS_LOST
		);
	} else if (SessionData.Wave == KFGRI.WaveMax) {
		class'SessionService'.static.GetInstance().UpdateStatus(
			SessionData.SessionId,
			SESSION_STATUS_WON
		);
	}
}

private function UploadWaveStats() {
	local PlayerStats PlayerData;
	local KFPlayerController C;
	local CreateWaveStatsBody Body;
	local PlayerReplicationInfo PRI;

	Body.SessionId = SessionData.SessionId;
	Body.Wave = SessionData.Wave;
	Body.Length = int(WI.RealTimeSeconds - SessionData.WaveStartedAt);

	if (SessionData.GameMode == 3) {
		Body.HasCDData = true;
		Body.CDData.SpawnCycle = class'CD_Utils'.static.GetSpawnCycle(WI);
		Body.CDData.MaxMonsters = class'CD_Utils'.static.GetMaxMonsters(WI);
		Body.CDData.WaveSizeFakes = class'CD_Utils'.static.GetWaveSizeFakes(WI);
		Body.CDData.ZedsType = class'CD_Utils'.static.GetZedsType(WI);
	}
	
	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;

		PlayerData = GetPlayerWaveStats(C);

		PRI = C.PlayerReplicationInfo;
		PlayerData.Uid = OS.UniqueNetIdToString(PRI.UniqueId);

		if (!C.bIsEosPlayer) {
			PlayerData.AuthId = OS.UniqueNetIdToInt64(PRI.UniqueId);
			PlayerData.AuthType = AT_STEAM;
		} else {
			PlayerData.AuthId = PlayerData.Uid;
			PlayerData.AuthType = AT_EGS;
		}

		if (PlayerData.Perk == 2) {
			PlayerData.ZedTimeLength = SessionData.ZedTimeDuration;
			PlayerData.ZedTimeCount = SessionData.ZedTimeCount;
		} else {
			PlayerData.ZedTimeLength = 0.0;
			PlayerData.ZedTimeCount = 0;
		}

		PlayerData.IsDead = (C.Pawn == None || !C.Pawn.IsAliveAndWell());

		Body.Players.AddItem(PlayerData);
	}

	class'StatsService'.static.GetInstance().CreateWaveStats(Body);
}


static function StatsRepo GetInstance() {
	local StatsRepo Instance;

	foreach Class'WorldInfo'.static.GetWorldInfo().DynamicActors(Class'StatsRepo', Instance) {      
		return Instance;        
	}

	return Instance;
}

function PlayerStats GetPlayerWaveStats(Controller C) {
	local string PlayerName;
	local PlayerStats stats;

	PlayerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	foreach Players(stats) {
		if (stats.PlayerName == PlayerName) {            
			return stats;
		}        
	}    

	return stats;
}

function AddZedKill(KFPlayerController C, name ZedKey) {
	local int i;
	local string playerName;
	local PlayerStats data;

	i = 0;
	playerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	while (i < Players.Length) {
		if (Players[i].PlayerName != playerName) {
			i++;
			continue;
		}

		switch (ZedKey) {
			case 'KFPawn_ZedClot_Cyst': 
				Players[i].Kills.Cyst++;
				return;
			case 'KFPawn_ZedClot_Alpha': 
				Players[i].Kills.AlphaClot++;
				return;
			case 'KFPawn_ZedClot_Slasher': 
				Players[i].Kills.Slasher++;
				return;
			case 'KFPawn_ZedCrawler': 
				Players[i].Kills.Crawler++;
				return;
			case 'KFPawn_ZedGorefast': 
				Players[i].Kills.Gorefast++;
				return;
			case 'KFPawn_ZedStalker': 
				Players[i].Kills.Stalker++;
				return;
			case 'KFPawn_ZedScrake': 
				Players[i].Kills.Scrake++;
				return;
			case 'KFPawn_ZedFleshpound': 
				Players[i].Kills.FP++;
				return;
			case 'KFPawn_ZedFleshpoundMini': 
				Players[i].Kills.QP++;
				return;
			case 'KFPawn_ZedBloat': 
				Players[i].Kills.Bloat++;
				return;
			case 'KFPawn_ZedSiren': 
				Players[i].Kills.Siren++;
				return;
			case 'KFPawn_ZedHusk': 
				Players[i].Kills.Husk++;
				return;
			case 'KFPawn_ZedClot_AlphaKing': 
				Players[i].Kills.Rioter++;
				return;
			case 'KFPawn_ZedCrawlerKing': 
				Players[i].Kills.EliteCrawler++;
				return;
			case 'KFPawn_ZedGorefastDualBlade': 
				Players[i].Kills.Gorefiend++;
				return;
			case 'KFPawn_ZedDAR_Emp': 
				Players[i].Kills.Edar++;
				return;
			case 'KFPawn_ZedDAR_Laser': 
				Players[i].Kills.Edar++;
				return;
			case 'KFPawn_ZedDAR_Rocket': 
				Players[i].Kills.Edar++;
				return;
			case 'KFPawn_ZedHans': 
				Players[i].Kills.Boss++;
				return;
			case 'KFPawn_ZedPatriarch': 
				Players[i].Kills.Boss++;
				return;
			case 'KFPawn_ZedFleshpoundKing': 
				Players[i].Kills.Boss++;
				return;
			case 'KFPawn_ZedBloatKing': 
				Players[i].Kills.Boss++;
				return;
			case 'KFPawn_ZedMatriarch': 
				Players[i].Kills.Boss++;
				return;
			default:
				return;
		}
	}

	if (playerName != "") {
		data.PlayerName = playerName;
		Players.AddItem(data);
		AddZedKill(C, ZedKey);
	}
}

function AddInjuredByZed(KFPlayerController C, name ZedKey, int Damage) {
	local int i;
	local string playerName;
	local PlayerStats data;

	if (Damage <= 0) return;

	i = 0;
	playerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	while (i < Players.Length) {
		if (Players[i].PlayerName != playerName) {
			i++;
			continue;
		}

		switch (ZedKey) {
			case 'KFPawn_ZedClot_Cyst': 
				Players[i].InjuredBy.Cyst += Damage;
				return;
			case 'KFPawn_ZedClot_Alpha': 
				Players[i].InjuredBy.AlphaClot += Damage;
				return;
			case 'KFPawn_ZedClot_Slasher': 
				Players[i].InjuredBy.Slasher += Damage;
				return;
			case 'KFPawn_ZedCrawler': 
				Players[i].InjuredBy.Crawler += Damage;
				return;
			case 'KFPawn_ZedGorefast': 
				Players[i].InjuredBy.Gorefast += Damage;
				return;
			case 'KFPawn_ZedStalker': 
				Players[i].InjuredBy.Stalker += Damage;
				return;
			case 'KFPawn_ZedScrake': 
				Players[i].InjuredBy.Scrake += Damage;
				return;
			case 'KFPawn_ZedFleshpound': 
				Players[i].InjuredBy.FP += Damage;
				return;
			case 'KFPawn_ZedFleshpoundMini': 
				Players[i].InjuredBy.QP += Damage;
				return;
			case 'KFPawn_ZedBloat': 
				Players[i].InjuredBy.Bloat += Damage;
				return;
			case 'KFPawn_ZedSiren': 
				Players[i].InjuredBy.Siren += Damage;
				return;
			case 'KFPawn_ZedHusk': 
				Players[i].InjuredBy.Husk += Damage;
				return;
			case 'KFPawn_ZedClot_AlphaKing': 
				Players[i].InjuredBy.Rioter += Damage;
				return;
			case 'KFPawn_ZedCrawlerKing': 
				Players[i].InjuredBy.EliteCrawler += Damage;
				return;
			case 'KFPawn_ZedGorefastDualBlade': 
				Players[i].InjuredBy.Gorefiend += Damage;
				return;
			case 'KFPawn_ZedDAR_Emp': 
				Players[i].InjuredBy.Edar += Damage;
				return;
			case 'KFPawn_ZedDAR_Laser': 
				Players[i].InjuredBy.Edar += Damage;
				return;
			case 'KFPawn_ZedDAR_Rocket': 
				Players[i].InjuredBy.Edar += Damage;
				return;
			case 'KFPawn_ZedHans': 
				Players[i].InjuredBy.Boss += Damage;
				return;
			case 'KFPawn_ZedPatriarch': 
				Players[i].InjuredBy.Boss += Damage;
				return;
			case 'KFPawn_ZedFleshpoundKing': 
				Players[i].InjuredBy.Boss += Damage;
				return;
			case 'KFPawn_ZedBloatKing': 
				Players[i].InjuredBy.Boss += Damage;
				return;
			case 'KFPawn_ZedMatriarch': 
				Players[i].InjuredBy.Boss += Damage;
				return;
			default:
				return;
		}
	}

	if (playerName != "") {
		data.PlayerName = playerName;
		Players.AddItem(data);
		AddZedKill(C, ZedKey);
	}
}

function AddEvent(KFPlayerController C, EventType type) {
	local int i;
	local string playerName;
	local PlayerStats data;

	i = 0;
	playerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	while (i < Players.Length) {
		if (Players[i].PlayerName != playerName) {
			i++;
			continue;
		}

		switch (type) {
			case ET_HUSK_BACKPACK:
				Players[i].HuskBackpackKills++;
				return;
			case ET_RAGED_BY_BP:
				Players[i].HuskRages++;
				return;
			default:
				return;
		}
	}

	if (playerName != "") {
		data.PlayerName = playerName;
		Players.AddItem(data);
		AddEvent(C, type);
	}
}

function UpdateNonKillStats(KFPlayerController C) {
	local int i;
	local string playerName;
	local PlayerStats data;

	i = 0;
	playerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	while (i < Players.Length) {
		if (Players[i].PlayerName != playerName) {
			i++;
			continue;
		}

		Players[i].ShotsFired += C.ShotsFired;
		Players[i].ShotsHit += C.ShotsHit;
		Players[i].ShotsHS = C.MatchStats.GetHeadShotsInWave();
		Players[i].DoshEarned = C.MatchStats.GetDoshEarnedInWave();
		Players[i].HealsGiven = C.MatchStats.GetHealGivenInWave();
		Players[i].HealsReceived = C.MatchStats.GetHealReceivedInWave();
		Players[i].DamageDealt = C.MatchStats.GetDamageDealtInWave();
		Players[i].DamageTaken = C.MatchStats.GetDamageTakenInWave();

		// `log("[UpdateNonKillStats]" @
		// 	"\nName="$C.GetHumanReadableName() @
		// 	"\nShotsFire="$PlayersStats[i].ShotsFired @
		// 	"ShotsHit="$PlayersStats[i].ShotsHit @
		// 	"ShotsHS="$PlayersStats[i].ShotsHS @
		// 	"\nDoshEarned="$PlayersStats[i].DoshEarned @
		// 	"\nHealsGiven="$PlayersStats[i].HealsGiven @
		// 	"HealsReceived="$PlayersStats[i].HealsReceived @
		// 	"\nDamageDealt="$PlayersStats[i].DamageDealt @
		// 	"DamageTaken="$PlayersStats[i].DamageTaken @
		// 	"\nCystKills="$PlayersStats[i].Kills.Cyst @
		// 	"AlphaClotKills="$PlayersStats[i].Kills.AlphaClot @
		// 	"SlasherKills="$PlayersStats[i].Kills.Slasher @
		// 	"StalkerKills="$PlayersStats[i].Kills.Stalker @
		// 	"CrawlerKills="$PlayersStats[i].Kills.Crawler @
		// 	"GorefastKills="$PlayersStats[i].Kills.Gorefast @
		// 	"\nRioterKills="$PlayersStats[i].Kills.Rioter @
		// 	"EliteCrawlerKills="$PlayersStats[i].Kills.EliteCrawler @
		// 	"GorefiendKills="$PlayersStats[i].Kills.Gorefiend @
		// 	"\nSirenKills="$PlayersStats[i].Kills.Siren @
		// 	"BloatKills="$PlayersStats[i].Kills.Bloat @
		// 	"EdarKills="$PlayersStats[i].Kills.Edar @
		// 	"\nHuskKills="$PlayersStats[i].Kills.Husk @
		// 	"HuskBackpackKills="$PlayersStats[i].HuskBackpackKills @
		// 	"HuskRages="$PlayersStats[i].HuskRages @
		// 	"\nScrakeKills="$PlayersStats[i].Kills.Scrake @
		// 	"FPKills="$PlayersStats[i].Kills.FP @
		// 	"QPKills="$PlayersStats[i].Kills.QP @
		// 	"BossKills="$PlayersStats[i].Kills.Boss
		// );

		return;
	}

	if (playerName != "") {
		data.PlayerName = playerName;
		Players.AddItem(data);
		UpdateNonKillStats(C);
	}
}

function ResetPlayerStats(KFPlayerController C) {
	local int i;
	local string playerName;
	local PlayerStats data;

	i = 0;
	playerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	while (i < Players.Length) {
		if (Players[i].PlayerName != playerName) {
			i++;
			continue;
		}

		Players[i].Perk = ConvertPerk(C.GetPerk().GetPerkClass());
		Players[i].Level = C.GetPerk().GetLevel();
		Players[i].Prestige = C.GetPerk().GetCurrentPrestigeLevel();

		Players[i].ShotsFired = -C.ShotsFired;
		Players[i].ShotsHit = -C.ShotsHit;
		Players[i].ShotsHS = 0;

		Players[i].DoshEarned = 0;
	
		Players[i].HealsGiven = 0;
		Players[i].HealsReceived = 0;

		Players[i].DamageDealt = 0;
		Players[i].DamageTaken = 0;

		Players[i].Kills.Cyst = 0;
		Players[i].Kills.AlphaClot = 0;
		Players[i].Kills.Slasher = 0;
		Players[i].Kills.Stalker = 0;
		Players[i].Kills.Crawler = 0;
		Players[i].Kills.Gorefast = 0;
		Players[i].Kills.Rioter = 0;
		Players[i].Kills.EliteCrawler = 0;
		Players[i].Kills.Gorefiend = 0;
		Players[i].Kills.Siren = 0;
		Players[i].Kills.Bloat = 0;
		Players[i].Kills.Edar = 0;
		Players[i].Kills.Husk = 0;
		Players[i].Kills.Scrake = 0;
		Players[i].Kills.FP = 0;
		Players[i].Kills.QP = 0;
		Players[i].Kills.Boss = 0;
		Players[i].Kills.Scrake = 0;
		Players[i].HuskBackpackKills = 0;
		Players[i].HuskRages = 0;

		Players[i].InjuredBy.Cyst = 0;
		Players[i].InjuredBy.AlphaClot = 0;
		Players[i].InjuredBy.Slasher = 0;
		Players[i].InjuredBy.Stalker = 0;
		Players[i].InjuredBy.Crawler = 0;
		Players[i].InjuredBy.Gorefast = 0;
		Players[i].InjuredBy.Rioter = 0;
		Players[i].InjuredBy.EliteCrawler = 0;
		Players[i].InjuredBy.Gorefiend = 0;
		Players[i].InjuredBy.Siren = 0;
		Players[i].InjuredBy.Bloat = 0;
		Players[i].InjuredBy.Edar = 0;
		Players[i].InjuredBy.Husk = 0;
		Players[i].InjuredBy.Scrake = 0;
		Players[i].InjuredBy.FP = 0;
		Players[i].InjuredBy.QP = 0;
		Players[i].InjuredBy.Boss = 0;
		Players[i].InjuredBy.Scrake = 0;

		return;
	}

	if (playerName != "") {
		data.PlayerName = playerName;
		Players.AddItem(data);
		ResetPlayerStats(C);
	}
}

private function UpdateGameData() {
	local UpdateGameDataRequest Body;

	if (SessionData.SessionId <= 0) return;

	Body.SessionId = SessionData.SessionId;

	Body.GameData.Wave = SessionData.Wave;
	Body.GameData.IsTraderTime = !SessionData.IsActive;
	Body.GameData.ZedsLeft = KFGRI.AIRemaining;
	Body.GameData.PlayersAlive = GetAliveCount();
	Body.GameData.PlayersOnline = GetConnectedCount();
	Body.GameData.MaxPlayers = WI.Game.MaxPlayersAllowed;

	if (SessionData.GameMode == 3) {
		Body.HasCDData = true;
		Body.CDData.SpawnCycle = class'CD_Utils'.static.GetSpawnCycle(WI);
		Body.CDData.MaxMonsters = class'CD_Utils'.static.GetMaxMonsters(WI);
		Body.CDData.WaveSizeFakes = class'CD_Utils'.static.GetWaveSizeFakes(WI);
		Body.CDData.ZedsType = class'CD_Utils'.static.GetZedsType(WI);
	}

	class'SessionService'.static.GetInstance().UpdateGameData(Body);
}

private function int ConvertPerk(Class<KFPerk> PerkClass) {
	if (PerkClass == class'KFPerk_Berserker')		return 1;
	if (PerkClass == class'KFPerk_Commando')		return 2;
	if (PerkClass == class'KFPerk_FieldMedic')		return 3;
	if (PerkClass == class'KFPerk_Sharpshooter')	return 4;
	if (PerkClass == class'KFPerk_Gunslinger')		return 5;
	if (PerkClass == class'KFPerk_Support')			return 6;
	if (PerkClass == class'KFPerk_Swat')			return 7;
	if (PerkClass == class'KFPerk_Demolitionist')	return 8;
	if (PerkClass == class'KFPerk_Firebug')			return 9;
	if (PerkClass == class'KFPerk_Survivalist')		return 10;

	return 0;
}

private function bool IsValidPlayer(KFPlayerController C) {
	return (
		C != None &&
		C.PlayerReplicationInfo != None &&
		!C.PlayerReplicationInfo.bOnlySpectator &&
		!C.PlayerReplicationInfo.bDemoOwner
	);
}

private function int GetAliveCount() {
	local int Count;
	local KFPlayerController C;

	Count = 0;
	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;

		if (C.Pawn != None && C.Pawn.IsAliveAndWell()) {
			Count++;
		}
	}

	return Count;
}

private function int GetConnectedCount() {
	local int Count;
	local KFPlayerController C;

	Count = 0;
	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;

		Count++;
	}

	return Count;
}