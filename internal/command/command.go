package command

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/n8nfelipe/backhere/internal/pathutil"
	"github.com/n8nfelipe/backhere/internal/shell"
	"github.com/n8nfelipe/backhere/internal/storage"
)

const usage = `BackHere — lembretes ligados ao diretório atual

Uso:
  backhere add <mensagem>
  backhere list
  backhere done <número>
  backhere clear
  backhere check [diretório]
  backhere init <bash|zsh|fish>`

type app struct {
	in          io.Reader
	out, errOut io.Writer
	dbPath      string
	getwd       func() (string, error)
}

func Run(args []string, in io.Reader, out, errOut io.Writer) int {
	dbPath, err := storage.DefaultDBPath()
	if err != nil {
		fmt.Fprintln(errOut, "backhere:", err)
		return 1
	}
	return run(app{in: in, out: out, errOut: errOut, dbPath: dbPath, getwd: os.Getwd}, args)
}

func run(a app, args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.out, usage)
		return 0
	}
	if args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		fmt.Fprintln(a.out, usage)
		return 0
	}
	if args[0] == "init" {
		if len(args) != 2 {
			return fail(a, "uso: backhere init <bash|zsh|fish>")
		}
		script, err := shell.Init(args[1])
		if err != nil {
			return fail(a, err.Error())
		}
		fmt.Fprintln(a.out, script)
		return 0
	}
	if !knownCommand(args[0]) {
		return fail(a, fmt.Sprintf("comando desconhecido: %s\n\n%s", args[0], usage))
	}
	if args[0] == "list" && len(args) != 1 {
		return fail(a, "uso: backhere list")
	}

	path, err := commandPath(a, args)
	if err != nil {
		return fail(a, err.Error())
	}
	if args[0] == "check" {
		if _, err := os.Stat(a.dbPath); os.IsNotExist(err) {
			return 0
		} else if err != nil {
			return fail(a, fmt.Sprintf("acessar banco: %v", err))
		}
	}
	var s *storage.Store
	if args[0] == "check" {
		s, err = storage.OpenExisting(a.dbPath)
	} else {
		s, err = storage.Open(a.dbPath)
	}
	if err != nil {
		return fail(a, err.Error())
	}
	defer s.Close()
	ctx := context.Background()

	switch args[0] {
	case "add":
		if len(args) < 2 {
			return fail(a, "uso: backhere add <mensagem>")
		}
		message := strings.TrimSpace(strings.Join(args[1:], " "))
		if message == "" {
			return fail(a, "a mensagem não pode estar vazia")
		}
		if err := s.Add(ctx, path, message); err != nil {
			return fail(a, err.Error())
		}
		fmt.Fprintln(a.out, "✓ Lembrete salvo aqui.")
	case "list", "check":
		reminders, err := s.Active(ctx, path)
		if err != nil {
			return fail(a, err.Error())
		}
		printReminders(a.out, reminders)
	case "done":
		if len(args) != 2 {
			return fail(a, "uso: backhere done <número>")
		}
		position, err := storage.ParsePosition(args[1])
		if err != nil {
			return fail(a, err.Error())
		}
		ok, err := s.Done(ctx, path, position)
		if err != nil {
			return fail(a, err.Error())
		}
		if !ok {
			return fail(a, "lembrete não encontrado")
		}
		fmt.Fprintln(a.out, "✓ Lembrete concluído.")
	case "clear":
		if len(args) != 1 {
			return fail(a, "uso: backhere clear")
		}
		reminders, err := s.Active(ctx, path)
		if err != nil {
			return fail(a, err.Error())
		}
		if len(reminders) == 0 {
			fmt.Fprintln(a.out, "Nenhum lembrete ativo aqui.")
			return 0
		}
		fmt.Fprintf(a.out, "Remover %d lembrete(s) ativo(s) deste diretório? [y/N] ", len(reminders))
		answer, _ := bufio.NewReader(a.in).ReadString('\n')
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" && answer != "s" && answer != "sim" {
			fmt.Fprintln(a.out, "Cancelado.")
			return 0
		}
		if _, err := s.Clear(ctx, path); err != nil {
			return fail(a, err.Error())
		}
		fmt.Fprintln(a.out, "✓ Lembretes removidos.")
	}
	return 0
}

func knownCommand(name string) bool {
	switch name {
	case "add", "list", "done", "clear", "check":
		return true
	default:
		return false
	}
}

func commandPath(a app, args []string) (string, error) {
	var raw string
	if args[0] == "check" && len(args) > 2 {
		return "", fmt.Errorf("uso: backhere check [diretório]")
	}
	if args[0] == "check" && len(args) == 2 {
		raw = args[1]
	} else {
		var err error
		raw, err = a.getwd()
		if err != nil {
			return "", fmt.Errorf("obter diretório atual: %w", err)
		}
	}
	return pathutil.NormalizePath(raw)
}

func printReminders(w io.Writer, reminders []storage.Reminder) {
	if len(reminders) == 0 {
		return
	}
	fmt.Fprintln(w, "📌 BackHere")
	fmt.Fprintln(w)
	for i, reminder := range reminders {
		fmt.Fprintf(w, "%d. %s\n", i+1, reminder.Message)
	}
}

func fail(a app, message string) int { fmt.Fprintln(a.errOut, "backhere:", message); return 1 }
