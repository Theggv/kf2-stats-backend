class IPServiceBase extends Object
	abstract;

delegate OnGetPublicIPCompleted(string Address);

function GetPublicIP();
