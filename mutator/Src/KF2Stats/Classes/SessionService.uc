class SessionService extends SessionServiceBase;

static function SessionService GetInstance() {
	local SessionService Instance;

	foreach Class'WorldInfo'.static.GetWorldInfo().DynamicActors(Class'SessionService', Instance) {      
		return Instance;        
	}

	return Instance;
}

public function CreateSession(
	int ServerId,
	int MapId,
	int Difficulty,
	int Length,
	int Mode
) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/sessions/",
		"POST",
		PrepareCreateSessionBody(ServerId, MapId, Difficulty, Length, Mode)
	);

	R.OnProcessRequestComplete = OnCreateSessionRequestComplete;
	R.ProcessRequest();
}

private function OnCreateSessionRequestComplete(
	HttpRequestInterface Request, 
	HttpResponseInterface Response, 
	bool bWasSuccessful
) {
	local int code, sessionId;
	local string content;
	local JsonObject parsedJson;

	code = 500;
	
	if (Response != None) {
		code = Response.GetResponseCode();
		content = Response.GetContentAsString();
	}

	if (code != 201) {
		`log("[CreateServer] failed with code:" @ code $ "." @ content);
		return;
	}

	parsedJson = class'JsonObject'.static.DecodeJson(content);
	sessionId = ParsedJson.GetIntValue("id");

	OnCreateSessionCompleted(sessionId);
}

public function UpdateStatus(int SessionId, int StatusId) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/sessions/status",
		"PUT",
		PrepareUpdateStatusBody(SessionId, StatusId)
	);

	R.ProcessRequest();
}

function UpdateGameData(UpdateGameDataRequest body) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/sessions/game-data",
		"PUT",
		PrepareUpdateGameDataBody(body)
	);

	R.ProcessRequest();
}


defaultproperties {
}