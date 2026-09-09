package shell

import (
	"strings"
	"testing"
)

func TestInitScripts(t *testing.T) {
	for _, name := range []string{"bash", "zsh", "fish"} {
		t.Run(name, func(t *testing.T) {
			script, err := Init(name)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(script, `backhere check "$PWD"`) {
				t.Fatalf("script does not check PWD: %s", script)
			}
			if !strings.Contains(script, "__backhere_last_pwd") {
				t.Fatalf("script has no session state: %s", script)
			}
		})
	}
}
