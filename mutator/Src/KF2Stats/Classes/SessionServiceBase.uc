class SessionServiceBase extends Info
	dependson (StatsServiceBase)
	abstract;

struct GameDataStruct {
	var int MaxPlayers;
	var int PlayersOnline;
	var int PlayersAlive;
	var int Wave;
	var bool IsTraderTime;
	var int ZedsLeft;
};

struct PlayerLiveData {
	var string PlayerName;
	var string AuthId;
	var AuthType AuthType;

	var int Perk;
	var int Level;
	var int Prestige;

	var int Health;
	var int Armor;

	var bool IsSpectator;

	structdefaultproperties {
		PlayerName = ""
		AuthId = ""
		AuthType = 0

		Perk = 0
		Level = 0
		Prestige = 0

		Health = 0
		Armor = 0

		IsSpectator = false
	}
};

struct UpdateGameDataRequest {
	var int SessionId;

	var GameDataStruct GameData;

	var bool HasCDData;
	var StatsServiceBase.CDStruct CDData;

	var array<PlayerLiveData> Players;
};

static function string PrepareCreateSessionBody( 
	int ServerId,
	int MapId,
	int Difficulty,
	int Length,
	int Mode
) {
	local JsonObject JSON;

	JSON = new Class'JsonObject';
	JSON.SetIntValue("server_id", ServerId);
	JSON.SetIntValue("map_id", MapId);
	JSON.SetIntValue("diff", Difficulty);
	JSON.SetIntValue("length", Length);
	JSON.SetIntValue("mode", Mode);

	return Class'JsonObject'.static.EncodeJson(JSON);
}

delegate OnCreateSessionCompleted(int sessionId);

function CreateSession(
	int ServerId,
	int MapId,
	int Difficulty,
	int Length,
	int Mode
);

static function string PrepareUpdateStatusBody( 
	int SessionId, int StatusId
) {
	local JsonObject JSON;

	JSON = new Class'JsonObject';
	JSON.SetIntValue("id", SessionId);
	JSON.SetIntValue("status", StatusId);

	return Class'JsonObject'.static.EncodeJson(JSON);
}

function UpdateStatus(int SessionId, int StatusId);

static function string PrepareUpdateGameDataBody(UpdateGameDataRequest body) {
	local PlayerLiveData D;
	local JsonObject Root, GameData, CDData, Players, PlayerData;

	Root = new Class'JsonObject';

	Root.SetIntValue("session_id", body.SessionId);

	GameData = new Class'JsonObject';
	GameData.SetIntValue("max_players", body.GameData.MaxPlayers);
	GameData.SetIntValue("players_online", body.GameData.PlayersOnline);
	GameData.SetIntValue("players_alive", body.GameData.PlayersAlive);
	GameData.SetIntValue("wave", body.GameData.Wave);
	GameData.SetBoolValue("is_trader_time", body.GameData.IsTraderTime);
	GameData.SetIntValue("zeds_left", body.GameData.ZedsLeft);
	Root.SetObject("game_data", GameData);

	if (body.HasCDData) {
		CDData = new Class'JsonObject';
		CDData.SetStringValue("spawn_cycle", body.CDData.SpawnCycle);
		CDData.SetIntValue("max_monsters", body.CDData.MaxMonsters);
		CDData.SetIntValue("wave_size_fakes", body.CDData.WaveSizeFakes);
		CDData.SetStringValue("zeds_type", body.CDData.ZedsType);
		Root.SetObject("cd_data", CDData);
	}

	Players = new Class'JsonObject';
	foreach Body.Players(D) {
		PlayerData = new Class'JsonObject';
		PlayerData.SetStringValue("name", D.PlayerName);
		PlayerData.SetStringValue("auth_id", D.AuthId);
		PlayerData.SetIntValue("auth_type", D.AuthType);

		PlayerData.SetIntValue("perk", D.Perk);
		PlayerData.SetIntValue("level", D.Level);
		PlayerData.SetIntValue("prestige", D.Prestige);

		PlayerData.SetIntValue("health", D.Health);
		PlayerData.SetIntValue("armor", D.Armor);

		PlayerData.SetBoolValue("is_spectator", D.IsSpectator);

		Players.ObjectArray.AddItem(PlayerData);
	}

	if (Body.Players.Length > 0) {
		Root.SetObject("players", Players);
	}

	return Class'JsonObject'.static.EncodeJson(Root);
}

function UpdateGameData(UpdateGameDataRequest body);
