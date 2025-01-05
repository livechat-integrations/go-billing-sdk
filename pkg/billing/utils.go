package billing

import (
	"fmt"
	"regexp"
)

var urlPattern = regexp.MustCompile(`^(.*)((?:livechat(?:inc)?|text)\.com)(/.*)?$`)

// EnvURL is a function that returns the URL with the environment
func EnvURL(url, lcEnv string) string {
	if lcEnv != "" {
		url = urlPattern.ReplaceAllString(url, fmt.Sprintf(`${1}%s.${2}${3}`, lcEnv))
	}
	return url
}
