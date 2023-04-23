package configure

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/invopop/yaml"
	_ "github.com/joho/godotenv/autoload"
)

// Load reads in the configuration file relative to the current path. Data
// is expected in YAML format with Golang template definitions.
func Load(file string, conf interface{}) error {
	f := path.Join(".", file)
	return parseConfigFile(f, conf)
}

func parseConfigFile(file string, conf interface{}) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	env, err := envToMap()
	if err != nil {
		return fmt.Errorf("reading environment: %w", err)
	}

	funcs := template.FuncMap{
		"indent": indent,
	}
	t, err := template.New("config").Funcs(funcs).Parse(string(data))
	if err != nil {
		return fmt.Errorf("parsing config template: %w", err)
	}

	buff := new(bytes.Buffer)
	if err := t.Execute(buff, env); err != nil {
		return fmt.Errorf("mapping values to template: %w", err)
	}

	if err := yaml.Unmarshal(buff.Bytes(), conf); err != nil {
		return fmt.Errorf("unmarshalling yaml: %w", err)
	}

	return nil
}

func envToMap() (map[string]string, error) {
	env := make(map[string]string)
	var err error

	for _, v := range os.Environ() {
		sv := strings.SplitN(v, "=", 2)
		env[sv[0]] = sv[1]
	}

	return env, err
}

// indent takes the string, finds all matching `\n`, and adds two
// spaces inmediatly after for each of the provided counts.
// This is useful for indenting environment variables to be correctly
// parsed by the YAML files.
func indent(text string, count int) string {
	spaces := ""
	for i := 0; i < count; i++ {
		spaces = spaces + "  "
	}
	return strings.ReplaceAll(text, "\n", "\n"+spaces)
}
