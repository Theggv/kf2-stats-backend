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

var array<StatsServiceBase.PlayerData> Players;
var SessionStruct SessionData;

var private WorldInfo WI;
var private KFGameInfo KFGI;
var private KFGameReplicationInfo KFGRI;
var private OnlineSubsystem OS;
var private MsgSpectator msgSpec;

static function StatsRepo GetInstance() {
	local StatsRepo Instance;

	foreach Class'WorldInfo'.static.GetWorldInfo().DynamicActors(Class'StatsRepo', Instance) {      
		return Instance;        
	}

	return Instance;
}

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

	msgSpec = Spawn(Class'MsgSpectator');

	SetTimer(0.1, true, 'UpdateLoop');
	SetTimer(15.0, true, 'UpdateGameData');
}

private function UpdateLoop() {
	local int currentWave;
	local bool isWaveActive;

	DetectPlayerDeath();

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

private function DetectPlayerDeath() {
    local KFPlayerController C;
	local int i;

	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;
		if (!GetPlayerStatsIndex(C, i)) continue;

		if (!Players[i].IsDead && C.Pawn != None && !C.Pawn.IsAliveAndWell()) {
			Players[i].IsDead = true;
		}
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
		
		ResetPlayerData(C);
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
	local int i;
	local KFPlayerController C;
	local CreateWaveStatsBody Body;

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
		if (!GetPlayerStatsIndex(C, i)) continue;

		if (Players[i].Perk == 2) {
			Players[i].Stats.ZedTimeLength = SessionData.ZedTimeDuration;
			Players[i].Stats.ZedTimeCount = SessionData.ZedTimeCount;
		} else {
			Players[i].Stats.ZedTimeLength = 0.0;
			Players[i].Stats.ZedTimeCount = 0;
		}

		Body.Players.AddItem(Players[i]);
	}

	class'StatsService'.static.GetInstance().CreateWaveStats(Body);

	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!IsValidPlayer(C)) continue;
		if (!GetPlayerStatsIndex(C, i)) continue;

		Players[i].Stats.RadioComms.RequestHealing = 0;
		Players[i].Stats.RadioComms.RequestDosh = 0;
		Players[i].Stats.RadioComms.RequestHelp = 0;
		Players[i].Stats.RadioComms.TauntZeds = 0;
		Players[i].Stats.RadioComms.FollowMe = 0;
		Players[i].Stats.RadioComms.GetToTheTrader = 0;
		Players[i].Stats.RadioComms.Affirmative = 0;
		Players[i].Stats.RadioComms.Negative = 0;
		Players[i].Stats.RadioComms.ThankYou = 0;
	}
}

function AddZedKill(KFPlayerController C, name ZedKey) {
	local int i;

	if (!GetPlayerStatsIndex(C, i)) return;

	switch (ZedKey) {
		case 'KFPawn_ZedClot_Cyst': 
			Players[i].Stats.Kills.Cyst++;
			return;
		case 'KFPawn_ZedClot_Alpha': 
			Players[i].Stats.Kills.AlphaClot++;
			return;
		case 'KFPawn_ZedClot_Slasher': 
			Players[i].Stats.Kills.Slasher++;
			return;
		case 'KFPawn_ZedCrawler': 
			Players[i].Stats.Kills.Crawler++;
			return;
		case 'KFPawn_ZedGorefast': 
			Players[i].Stats.Kills.Gorefast++;
			return;
		case 'KFPawn_ZedStalker': 
			Players[i].Stats.Kills.Stalker++;
			return;
		case 'KFPawn_ZedScrake': 
			Players[i].Stats.Kills.Scrake++;
			return;
		case 'KFPawn_ZedFleshpound': 
			Players[i].Stats.Kills.FP++;
			return;
		case 'KFPawn_ZedFleshpoundMini': 
			Players[i].Stats.Kills.QP++;
			return;
		case 'KFPawn_ZedBloat': 
			Players[i].Stats.Kills.Bloat++;
			return;
		case 'KFPawn_ZedSiren': 
			Players[i].Stats.Kills.Siren++;
			return;
		case 'KFPawn_ZedHusk': 
			Players[i].Stats.Kills.Husk++;
			return;
		case 'KFPawn_ZedClot_AlphaKing': 
			Players[i].Stats.Kills.Rioter++;
			return;
		case 'KFPawn_ZedCrawlerKing': 
			Players[i].Stats.Kills.EliteCrawler++;
			return;
		case 'KFPawn_ZedGorefastDualBlade': 
			Players[i].Stats.Kills.Gorefiend++;
			return;
		case 'KFPawn_ZedDAR_Emp': 
			Players[i].Stats.Kills.Edar++;
			return;
		case 'KFPawn_ZedDAR_Laser': 
			Players[i].Stats.Kills.Edar++;
			return;
		case 'KFPawn_ZedDAR_Rocket': 
			Players[i].Stats.Kills.Edar++;
			return;
		case 'KFPawn_ZedHans': 
			Players[i].Stats.Kills.Boss++;
			return;
		case 'KFPawn_ZedPatriarch': 
			Players[i].Stats.Kills.Boss++;
			return;
		case 'KFPawn_ZedFleshpoundKing': 
			Players[i].Stats.Kills.Boss++;
			return;
		case 'KFPawn_ZedBloatKing': 
			Players[i].Stats.Kills.Boss++;
			return;
		case 'KFPawn_ZedMatriarch': 
			Players[i].Stats.Kills.Boss++;
			return;
		default:
			Players[i].Stats.Kills.Custom++;
			return;
	}
}

function AddEvent(KFPlayerController C, EventType type) {
	local int i;

	if (!GetPlayerStatsIndex(C, i)) return;

	switch (type) {
		case ET_HUSK_BACKPACK:
			Players[i].Stats.HuskBackpackKills++;
			return;
		case ET_RAGED_BY_BP:
			Players[i].Stats.HuskRages++;
			return;
		default:
			return;
	}
}

function AddRadioComms(PlayerReplicationInfo PRI, int Type) {
	local int i;

	if (!GetPlayerStatsByPRI(KFPlayerReplicationInfo(PRI), i)) return;

	switch (Type) {
		case 0:
			Players[i].Stats.RadioComms.RequestHealing++;
			return;
		case 1:
			Players[i].Stats.RadioComms.RequestDosh++;
			return;
		case 2:
			Players[i].Stats.RadioComms.RequestHelp++;
			return;
		case 3:
			Players[i].Stats.RadioComms.TauntZeds++;
			return;
		case 4:
			Players[i].Stats.RadioComms.FollowMe++;
			return;
		case 5:
			Players[i].Stats.RadioComms.GetToTheTrader++;
			return;
		case 6:
			Players[i].Stats.RadioComms.Affirmative++;
			return;
		case 7:
			Players[i].Stats.RadioComms.Negative++;
			return;
		case 9:
			Players[i].Stats.RadioComms.ThankYou++;
			return;
	}
}

function UpdateNonKillStats(KFPlayerController C) {
	local int i;

	if (!GetPlayerStatsIndex(C, i)) return;

	Players[i].Stats.ShotsFired += C.ShotsFired;
	Players[i].Stats.ShotsHit += C.ShotsHit;
	Players[i].Stats.ShotsHS = C.MatchStats.GetHeadShotsInWave();
	Players[i].Stats.DoshEarned = C.MatchStats.GetDoshEarnedInWave();
	Players[i].Stats.HealsGiven = C.MatchStats.GetHealGivenInWave();
	Players[i].Stats.HealsReceived = C.MatchStats.GetHealReceivedInWave();
	Players[i].Stats.DamageDealt = C.MatchStats.GetDamageDealtInWave();
	Players[i].Stats.DamageTaken = C.MatchStats.GetDamageTakenInWave();
}

function ResetPlayerData(KFPlayerController C) {
	local int i;
	local PlayerReplicationInfo PRI;

	if (!GetPlayerStatsIndex(C, i)) return;

	PRI = C.PlayerReplicationInfo;
	Players[i].UniqueId = OS.UniqueNetIdToString(PRI.UniqueId);

	Players[i].Perk = ConvertPerk(C.GetPerk().GetPerkClass());
	Players[i].Level = C.GetPerk().GetLevel();
	Players[i].Prestige = C.GetPerk().GetCurrentPrestigeLevel();
	Players[i].IsDead = false;

	Players[i].Stats.ShotsFired = -C.ShotsFired;
	Players[i].Stats.ShotsHit = -C.ShotsHit;
	Players[i].Stats.ShotsHS = 0;

	Players[i].Stats.DoshEarned = 0;

	Players[i].Stats.HealsGiven = 0;
	Players[i].Stats.HealsReceived = 0;

	Players[i].Stats.DamageDealt = 0;
	Players[i].Stats.DamageTaken = 0;

	Players[i].Stats.Kills.Cyst = 0;
	Players[i].Stats.Kills.AlphaClot = 0;
	Players[i].Stats.Kills.Slasher = 0;
	Players[i].Stats.Kills.Stalker = 0;
	Players[i].Stats.Kills.Crawler = 0;
	Players[i].Stats.Kills.Gorefast = 0;
	Players[i].Stats.Kills.Rioter = 0;
	Players[i].Stats.Kills.EliteCrawler = 0;
	Players[i].Stats.Kills.Gorefiend = 0;
	Players[i].Stats.Kills.Siren = 0;
	Players[i].Stats.Kills.Bloat = 0;
	Players[i].Stats.Kills.Edar = 0;
	Players[i].Stats.Kills.Husk = 0;
	Players[i].Stats.Kills.Scrake = 0;
	Players[i].Stats.Kills.FP = 0;
	Players[i].Stats.Kills.QP = 0;
	Players[i].Stats.Kills.Boss = 0;
	Players[i].Stats.Kills.Custom = 0;
	Players[i].Stats.HuskBackpackKills = 0;
	Players[i].Stats.HuskRages = 0;
}

private function UpdateGameData() {
	local KFPlayerController C;
	local SessionServiceBase.UpdateGameDataRequest Body;
	local SessionServiceBase.PlayerLiveData PLiveData;

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

	foreach WI.AllControllers(Class'KFPlayerController', C) {
		if (!GetPlayerLiveData(C, PLiveData)) continue;

		Body.Players.AddItem(PLiveData);
	}

	class'SessionService'.static.GetInstance().UpdateGameData(Body);
}

private function bool GetPlayerLiveData(
	KFPlayerController C,
	out SessionServiceBase.PlayerLiveData OutData
) {
	local string UniqueId, PlayerName;
	local KFPlayerReplicationInfo PRI;
	local PlayerLiveData Data;

	if (C == None) return false;

	PRI = KFPlayerReplicationInfo(C.PlayerReplicationInfo);
	if (PRI == None) return false;

	UniqueId = OS.UniqueNetIdToString(PRI.UniqueId);
	PlayerName = KFPlayerReplicationInfo(C.PlayerReplicationInfo).PlayerName;

	Data.PlayerName = PlayerName;

	if (!C.bIsEosPlayer) {
		Data.AuthId = OS.UniqueNetIdToInt64(PRI.UniqueId);
		Data.AuthType = AT_STEAM;
	} else {
		Data.AuthId = UniqueId;
		Data.AuthType = AT_EGS;
	}
	
	if (PRI.bOnlySpectator || PRI.bDemoOwner) {
		Data.IsSpectator = true;
	}

	if (C.Pawn != None && KFPawn_Human(C.Pawn) != None) {
		Data.Perk = ConvertPerk(C.GetPerk().GetPerkClass());
		Data.Level = C.GetPerk().GetLevel();
		Data.Prestige = C.GetPerk().GetCurrentPrestigeLevel();

		Data.Health = KFPawn_Human(C.Pawn).Health;
		Data.Armor = KFPawn_Human(C.Pawn).Armor;
	}

	OutData = Data;

	return true;
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

private function bool GetPlayerStatsByPRI(
	KFPlayerReplicationInfo PRI,
	optional out int Index
) {
	local string UniqueId, PlayerName;
	local StatsServiceBase.PlayerData Iter;

	if (PRI == None) return false;

	UniqueId = OS.UniqueNetIdToString(PRI.UniqueId);
	PlayerName = PRI.PlayerName;
	Index = 0;

	foreach Players(Iter) {
		if (UniqueId == Iter.UniqueId) {
			return true;
		}

		Index++;
	}

	Iter.UniqueId = UniqueId;
	Iter.PlayerName = PlayerName;

	if (!PRI.KFPlayerOwner.bIsEosPlayer) {
		Iter.AuthId = OS.UniqueNetIdToInt64(PRI.UniqueId);
		Iter.AuthType = AT_STEAM;
	} else {
		Iter.AuthId = Iter.UniqueId;
		Iter.AuthType = AT_EGS;
	}

	Players.AddItem(Iter);
	Index = Players.Length - 1;

	return true;
}

private function bool GetPlayerStatsIndex(
	KFPlayerController C, 
	optional out int Index
) {
	if (C == None) return false;

	return GetPlayerStatsByPRI(KFPlayerReplicationInfo(C.PlayerReplicationInfo), Index);
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

