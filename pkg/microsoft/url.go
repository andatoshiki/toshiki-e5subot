package microsoft

import (
	"fmt"
	"net/url"
)

const (
	apiURL   string = "https://login.microsoftonline.com"
	graphURL string = "https://graph.microsoft.com"
	redirect string = "http://localhost/toshiki-e5subot"
	scope    string = "openid offline_access mail.read user.read"
)

func GetAuthURL(clientID string) string {
	return fmt.Sprintf(
		"https://telegra.ph/%E4%BF%8A%E6%A8%B9%E3%81%AEE5subot%E6%95%99%E7%A8%8B-07-31",
		url.QueryEscape(redirect),
		url.QueryEscape(scope),
	)
}

func GetRegURL() string {
	ru := "https://developer.microsoft.com/en-us/graph/quick-start?appID=_appId_&appName=_appName_&redirectUrl=http://localhost:8000&platform=option-windowsuniversal"
	deeplink := fmt.Sprintf("/quickstart/graphIO?publicClientSupport=false&appName=toshiki-e5subot&redirectUrl=%s&allowImplicitFlow=false&ru=%s", redirect, url.QueryEscape(ru))
	appUrl := fmt.Sprintf("https://apps.dev.microsoft.com/?deepLink=%s", url.QueryEscape(deeplink))
	return appUrl
}
