class ServerService extends ServerServiceBase;

public function CreateServer(string ServerName, string ServerAddress) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/servers/",
		"POST",
		PrepareCreateServerBody(ServerName, ServerAddress)
	);

	R.OnProcessRequestComplete = OnCreateServerRequestComplete;
	R.ProcessRequest();
}

private function OnCreateServerRequestComplete(
	HttpRequestInterface Request, 
	HttpResponseInterface Response, 
	bool bWasSuccessful
) {
	local int code, serverId;
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
	serverId = ParsedJson.GetIntValue("id");

	OnCreateServerCompleted(serverId);
}

defaultproperties {
}