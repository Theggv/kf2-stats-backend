class SessionServiceBase extends Info
	abstract;

struct CDStruct {
	var string SpawnCycle;
	var int MaxMonsters;
	var int WaveSizeFakes;
	var string ZedsType;
};

struct GameDataStruct {
	var int MaxPlayers;
	var int PlayersOnline;
	var int PlayersAlive;
	var int Wave;
	var bool IsTraderTime;
	var int ZedsLeft;
};

struct UpdateGameDataRequest {
	var int SessionId;

	var GameDataStruct GameData;

	var bool HasCDData;
	var CDStruct CDData;
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
	local JsonObject Json, JsonGameData, JsonCDData;

	Json = new Class'JsonObject';

	Json.SetIntValue("session_id", body.SessionId);

	JsonGameData = new Class'JsonObject';
	JsonGameData.SetIntValue("max_players", body.GameData.MaxPlayers);
	JsonGameData.SetIntValue("players_online", body.GameData.PlayersOnline);
	JsonGameData.SetIntValue("players_alive", body.GameData.PlayersAlive);
	JsonGameData.SetIntValue("wave", body.GameData.Wave);
	JsonGameData.SetBoolValue("is_trader_time", body.GameData.IsTraderTime);
	JsonGameData.SetIntValue("zeds_left", body.GameData.ZedsLeft);
	Json.SetObject("game_data", JsonGameData);

	if (body.HasCDData) {
		JsonCDData = new Class'JsonObject';
		JsonCDData.SetStringValue("spawn_cycle", body.CDData.SpawnCycle);
		JsonCDData.SetIntValue("max_monsters", body.CDData.MaxMonsters);
		JsonCDData.SetIntValue("wave_size_fakes", body.CDData.WaveSizeFakes);
		JsonCDData.SetStringValue("zeds_type", body.CDData.ZedsType);
		Json.SetObject("cd_data", JsonCDData);
	}

	return Class'JsonObject'.static.EncodeJson(Json);
}

function UpdateGameData(UpdateGameDataRequest body);
