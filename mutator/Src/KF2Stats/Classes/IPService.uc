class IPService extends IPServiceBase;

public function GetPublicIP() {
	local HttpRequestInterface R;

	R = class'HttpUtils'.static.PrepareRequest(
		"https://api.ipify.org/?format=json",
		"GET"
	);

	R.OnProcessRequestComplete = OnGetPublicIPRequestComplete;
	R.ProcessRequest();
}

private function OnGetPublicIPRequestComplete(
	HttpRequestInterface Request, 
	HttpResponseInterface Response, 
	bool bWasSuccessful
) {
	local int code;
	local string content, address;
	local JsonObject parsedJson;

	code = 500;

	if (Response != None) {
		code = Response.GetResponseCode();
		content = Response.GetContentAsString();
	}

	if (code != 200) {
		`log("[GetPublicIP] failed, code:" @ code $ "." @ content);
		return;
	}

	parsedJson = class'JsonObject'.static.DecodeJson(content);
	address = ParsedJson.GetStringValue("ip");

	OnGetPublicIPCompleted(address);
}

defaultproperties {
}