class ServerServiceBase extends Object
	abstract;

static function string PrepareCreateServerBody(string ServerName, string ServerAddress) {
	local JsonObject JSON;

	JSON = new Class'JsonObject';
	JSON.SetStringValue("name", ServerName);
	JSON.SetStringValue("address", ServerAddress);

	return Class'JsonObject'.static.EncodeJson(JSON);
}

delegate OnCreateServerCompleted(int serverId);

function CreateServer(string ServerName, string ServerAddress);
