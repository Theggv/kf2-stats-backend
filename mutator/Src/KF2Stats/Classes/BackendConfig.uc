class BackendConfig extends Object
	config(KF2Stats);

/** Base backend url */
var public config string BaseUrl;
var public config string SecretToken;

public static function InitConfig(int Version, int LatestVersion) {
	switch (Version) {
		case `NO_CONFIG:
			ApplyDefault();
		default: 
			break;
	}

	if (Version != LatestVersion) {
		StaticSaveConfig();
	}
}

private static function ApplyDefault() {
	default.BaseUrl = "http://localhost:3000";
	default.SecretToken = "";
}

defaultproperties {
}