package command

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type harness struct {
	app         app
	out, errOut *bytes.Buffer
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dir := t.TempDir()
	out, errOut := &bytes.Buffer{}, &bytes.Buffer{}
	return &harness{app: app{in: strings.NewReader(""), out: out, errOut: errOut,
		dbPath: filepath.Join(t.TempDir(), "backhere.db"), getwd: func() (string, error) { return dir, nil }}, out: out, errOut: errOut}
}

func (h *harness) call(args ...string) (int, string, string) {
	h.out.Reset()
	h.errOut.Reset()
	code := run(h.app, args)
	return code, h.out.String(), h.errOut.String()
}

func TestCommandsLifecycle(t *testing.T) {
	h := newHarness(t)
	if code, out, _ := h.call("check"); code != 0 || out != "" {
		t.Fatalf("empty check = %d, %q", code, out)
	}
	if code, _, errOut := h.call("add", "verificar", "docker", "compose"); code != 0 {
		t.Fatalf("add = %d, %s", code, errOut)
	}
	if code, out, _ := h.call("list"); code != 0 || !strings.Contains(out, "1. verificar docker compose") {
		t.Fatalf("list = %d, %q", code, out)
	}
	if code, _, errOut := h.call("done", "1"); code != 0 {
		t.Fatalf("done = %d, %s", code, errOut)
	}
	if _, out, _ := h.call("check"); out != "" {
		t.Fatalf("check after done = %q", out)
	}
}

func TestClearConfirmation(t *testing.T) {
	h := newHarness(t)
	h.call("add", "one")
	h.call("add", "two")
	h.app.in = strings.NewReader("sim\n")
	if code, out, errOut := h.call("clear"); code != 0 || !strings.Contains(out, "Lembretes removidos") {
		t.Fatalf("clear = %d, %q, %q", code, out, errOut)
	}
	if _, out, _ := h.call("list"); out != "" {
		t.Fatalf("list after clear = %q", out)
	}
}

func TestCheckExplicitDirectoryIsolation(t *testing.T) {
	h := newHarness(t)
	dir2 := t.TempDir()
	h.call("add", "only here")
	if _, out, _ := h.call("check", dir2); out != "" {
		t.Fatalf("other directory = %q", out)
	}
}

func TestInitDoesNotCreateDatabase(t *testing.T) {
	h := newHarness(t)
	if code, out, _ := h.call("init", "bash"); code != 0 || !strings.Contains(out, "PROMPT_COMMAND") {
		t.Fatalf("init = %d, %q", code, out)
	}
}

func TestEmptyCheckDoesNotCreateDatabase(t *testing.T) {
	h := newHarness(t)
	if code, out, errOut := h.call("check"); code != 0 || out != "" || errOut != "" {
		t.Fatalf("check = %d, %q, %q", code, out, errOut)
	}
	if _, err := os.Stat(h.app.dbPath); !os.IsNotExist(err) {
		t.Fatalf("check created database: %v", err)
	}
}
