class KF2Stats extends Info
	dependson(ServerServiceBase)
    config(KF2Stats);

const LatestVersion = 3;

var public config int			Version;
var public config E_LogLevel	LogLevel;

var private OnlineSubsystem 	OS;
var StatsRepo 					Stats;

const ServerServiceClass = 		class'ServerService';
const MapsServiceClass = 		class'MapsService';
const SessionServiceClass = 	class'SessionService';
const IPServiceClass =			class'IPService';
const BackendConfigClass = 		class'BackendConfig';

public simulated function bool SafeDestroy() {
	`Log_Trace();

	return (bPendingDelete || bDeleteMe || Destroy());
}

public event PreBeginPlay() {
	`Log_Trace();

	if (WorldInfo.NetMode == NM_Client) {
		`Log_Fatal("Wrong NetMode:" @ WorldInfo.NetMode);
		SafeDestroy();
		return;
	}

	Stats = Spawn(Class'StatsRepo');

	Super.PreBeginPlay();

	PreInit();
}

private function PreInit() {
	if (Version == `NO_CONFIG) {
		LogLevel = LL_Info;
		SaveConfig();
	}

	BackendConfigClass.static.InitConfig(Version, LatestVersion);

	if (LatestVersion != Version) {
		Version = LatestVersion;
		SaveConfig();
	}

	OS = class'GameEngine'.static.GetOnlineSubsystem();
}

public event PostBeginPlay() {
	`Log_Trace();

	if (bPendingDelete || bDeleteMe) return;

	Super.PostBeginPlay();

	PostInit();
}

function PostInit() {
	`Log_Trace();

	if (WorldInfo.Game == None || WorldInfo.GRI == None) {
		SetTimer(0.2, false, nameof(PostInit));
		return;
	}

	Spawn(class'SessionService');
	Spawn(class'StatsService');

	GetServerAddress();
}

function CreateServer(string Address) {
	local ServerService service;

	service = new ServerServiceClass;
	service.OnCreateServerCompleted = OnCreateServerCompleted;

	service.CreateServer(GetServerName(), Address);
}

function CreateMap() {
	local string mapName;
	local MapsService service;

	mapName = Caps(WorldInfo.GetMapName(true));

	service = new MapsServiceClass;
	service.OnCreateMapCompleted = OnCreateMapCompleted;

	service.CreateMap(mapName);
}

function CreateSession() {
	local SessionService service;

	Stats.SessionData.GameDifficulty = GetGameDifficulty();
	Stats.SessionData.GameLength = GetGameLength();
	Stats.SessionData.GameMode = GetGameMode();

	service = SessionServiceClass.static.GetInstance();
	service.OnCreateSessionCompleted = OnCreateSessionCompleted;

	service.CreateSession(
		Stats.SessionData.ServerId, 
		Stats.SessionData.MapId, 
		Stats.SessionData.GameDifficulty, 
		Stats.SessionData.GameLength, 
		Stats.SessionData.GameMode
	);
}

private function string GetServerName() {
	return WorldInfo.Game.GameReplicationInfo.ServerName;
}

private function GetServerAddress() {
	local IPService service;

	service = new IPServiceClass;
	service.OnGetPublicIPCompleted = OnGetPublicIPCompleted;
	service.GetPublicIP();
}

private function OnGetPublicIPCompleted(string Address) {
	local string AddressUrl;
	local array<string> UrlParts;

	AddressUrl = WorldInfo.GetAddressURL();
	
	if (InStr(AddressUrl, ":") > INDEX_NONE) {
		UrlParts = SplitString(AddressUrl, ":", false);
	}

	UrlParts[0] = Address;

	if (UrlParts[1] == "") {
		UrlParts[1] = "7777";
	}

	Address = UrlParts[0] $ ":" $ UrlParts[1];

	CreateServer(Address);
}

private function OnCreateServerCompleted(int Id) {
	Stats.SessionData.ServerId = Id;
	CreateMap();
}

private function OnCreateMapCompleted(int Id) {
	Stats.SessionData.MapId = Id;
	CreateSession();
}
private function OnCreateSessionCompleted(int Id) {
	Stats.SessionData.SessionId = Id;
}

private function int GetGameDifficulty() {
	return int(WorldInfo.Game.GameDifficulty) + 1;
}

private function int GetGameLength() {
	return KFGameInfo(WorldInfo.Game).MyKFGRI.WaveMax - 1;
}

private function int GetGameMode() {
	local KFGameInfo KFGI;

	KFGI = KFGameInfo(WorldInfo.Game);

	if (KFGI == None) return 0;

	// CD Support
	if (KFGI.IsA('CD_Survival')) return 3;

	// Default modes
	if (KFGameInfo_Endless(KFGI) != None) return 2;
	if (KFGameInfo_WeeklySurvival(KFGI) != None) return 4;
	if (KFGameInfo_Objective(KFGI) != None) return 5;
	if (KFGameInfo_VersusSurvival(KFGI) != None) return 6;

	if (KFGameInfo_Survival(KFGI) != None) return 1;

	return 0;
}

defaultproperties {
}