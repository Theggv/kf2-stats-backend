package steamapi

import (
	"strconv"
	"strings"
)

func tryGetSteamId(params string) (string, bool) {
	parts := strings.SplitSeq(params, "&")

	for part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		key := keyValue[0]
		value := keyValue[1]

		if key == "openid.claimed_id" {
			valueParts := strings.Split(value, "%2F")
			if len(valueParts) < 2 {
				return "", false
			}

			steamId := valueParts[len(valueParts)-1]
			_, err := strconv.ParseFloat(steamId, 64)
			if err != nil {
				return "", false
			}

			return steamId, true
		}
	}

	return "", false
}
