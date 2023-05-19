package configure_test

import (
	"os"
	"testing"

	"github.com/invopop/configure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ExampleConf struct {
	User struct {
		Type       string `json:"type"`
		Name       string `json:"name"`
		FromDotenv string `json:"from_dotenv"`
		Default    string `json:"default"`
		DefaultOr  string `json:"default_or"`
		Embed      string `json:"embed"`
		EmbedIf    string `json:"embed_if"`
	} `json:"user_test"`
}

func TestYAMLConfig(t *testing.T) {
	require.NoError(t, os.Setenv("EMBEDDED", "sample\ntext"))
	conf := new(ExampleConf)
	err := configure.Load("samples/config.yaml.tmpl", conf)
	assert.NoError(t, err)

	assert.Equal(t, os.Getenv("USER"), conf.User.Name, "expected name to be parsed correctly")
	assert.Equal(t, "foobar", conf.User.FromDotenv, "expected data to be loaded from .env file")
	assert.Equal(t, "bar", conf.User.Default, "expected default value to be parsed correctly")
	assert.Equal(t, "bar", conf.User.DefaultOr, "expected default value to be parsed correctly")
	assert.Equal(t, "sample\ntext", conf.User.Embed, "expected embedded value to be parsed correctly")
	assert.Equal(t, "no content", conf.User.EmbedIf, "expected embedded if value to be parsed correctly")
}
