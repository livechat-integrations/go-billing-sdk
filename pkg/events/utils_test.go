package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvURL(t *testing.T) {
	t.Run("labs", func(t *testing.T) {
		assert.Equal(t, "https://labs.livechatinc.com/", EnvURL("https://livechatinc.com/", "labs"))
	})
}
