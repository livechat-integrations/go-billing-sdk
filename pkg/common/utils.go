package common

import (
	"fmt"
	"regexp"

	"github.com/rs/xid"
)

var urlPattern = regexp.MustCompile(`^(.*)((?:livechat(?:inc)?|text)\.com)(/.*)?$`)

// EnvURL is a function that returns the URL with the environment
func EnvURL(url, lcEnv string) string {
	if lcEnv != "" {
		url = urlPattern.ReplaceAllString(url, fmt.Sprintf(`${1}%s.${2}${3}`, lcEnv))
	}
	return url
}

type XIdProviderInterface interface {
	GenerateXId() string
}

type XIdProvider struct {
}

func (XIdProvider) GenerateXId() string {
	return xid.New().String()
}
