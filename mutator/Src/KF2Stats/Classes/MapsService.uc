class MapsService extends MapsServiceBase;

public function CreateMap(string MapName) {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		class'BackendConfig'.default.BaseUrl $ "/api/maps/",
		"POST",
		PrepareCreateMapBody(MapName)
	);

	R.OnProcessRequestComplete = OnCreateMapRequestComplete;
	R.ProcessRequest();
}

private function OnCreateMapRequestComplete(
	HttpRequestInterface Request, 
	HttpResponseInterface Response, 
	bool bWasSuccessful
) {
	local int code, mapId;
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
	mapId = ParsedJson.GetIntValue("id");

	OnCreateMapCompleted(mapId);
}

defaultproperties {
}