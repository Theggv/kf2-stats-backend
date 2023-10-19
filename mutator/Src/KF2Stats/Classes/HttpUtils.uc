class HttpUtils extends Object;

static function HttpRequestInterface PrepareRequest(
	string url, 
	string requestType, 
	optional string payload
) {
	local HttpRequestInterface R;

	R = Class'HttpFactory'.static.CreateRequest();
	R.SetURL(url);
	R.SetVerb(requestType);
	R.SetContentAsString(payload);
	R.SetHeader("Authorization", "Bearer" @ class'BackendConfig'.default.SecretToken);

	return R;
}