class MapsServiceBase extends Object
	abstract;

static function string PrepareCreateMapBody(string MapName) {
	local JsonObject JSON;

	JSON = new Class'JsonObject';
	JSON.SetStringValue("name", MapName);
	JSON.SetStringValue("preview", "");

	return Class'JsonObject'.static.EncodeJson(JSON);
}

delegate OnCreateMapCompleted(int sessionId);

function CreateMap(string MapName);
