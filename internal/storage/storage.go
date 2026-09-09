package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	_ "modernc.org/sqlite"
)

type Reminder struct {
	ID        int64
	Message   string
	CreatedAt time.Time
}

type Store struct{ db *sql.DB }

func DefaultDBPath() (string, error) {
	var base string
	var err error
	directory := "BackHere"
	switch runtime.GOOS {
	case "linux":
		directory = "backhere"
		base = os.Getenv("XDG_DATA_HOME")
		if base == "" {
			var home string
			home, err = os.UserHomeDir()
			base = filepath.Join(home, ".local", "share")
		}
	case "windows":
		base = os.Getenv("LOCALAPPDATA")
		if base == "" {
			base, err = os.UserConfigDir()
		}
	default:
		base, err = os.UserConfigDir()
	}
	if err != nil {
		return "", fmt.Errorf("localizar diretório de dados: %w", err)
	}
	if base == "" {
		return "", errors.New("diretório de dados não encontrado")
	}
	return filepath.Join(base, directory, "backhere.db"), nil
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("criar diretório de dados: %w", err)
	}
	s, err := open(path, false)
	if err != nil {
		return nil, err
	}
	if err := s.migrate(context.Background()); err != nil {
		_ = s.Close()
		return nil, err
	}
	_ = os.Chmod(path, 0o600)
	return s, nil
}

// OpenExisting skips schema setup and is used by the latency-sensitive check
// command. mode=rw guarantees that a check never creates a database.
func OpenExisting(path string) (*Store, error) {
	return open(path, true)
}

func open(path string, existingOnly bool) (*Store, error) {
	q := url.Values{}
	q.Set("_busy_timeout", "3000")
	q.Set("_synchronous", "NORMAL")
	if existingOnly {
		q.Set("mode", "rw")
	} else {
		q.Set("_journal_mode", "WAL")
	}
	db, err := sql.Open("sqlite", path+"?"+q.Encode())
	if err != nil {
		return nil, fmt.Errorf("abrir banco: %w", err)
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate(ctx context.Context) error {
	const schema = `
CREATE TABLE IF NOT EXISTS reminders (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 path TEXT NOT NULL,
 message TEXT NOT NULL,
 created_at TEXT NOT NULL,
 completed_at TEXT
);
CREATE INDEX IF NOT EXISTS reminders_active_path_idx
 ON reminders(path, completed_at, id);`
	if _, err := s.db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("preparar banco: %w", err)
	}
	return nil
}

func (s *Store) Add(ctx context.Context, path, message string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO reminders(path, message, created_at) VALUES (?, ?, ?)`,
		path, message, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("salvar lembrete: %w", err)
	}
	return nil
}

func (s *Store) Active(ctx context.Context, path string) ([]Reminder, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, message, created_at FROM reminders
         WHERE path = ? AND completed_at IS NULL ORDER BY id`, path)
	if err != nil {
		return nil, fmt.Errorf("listar lembretes: %w", err)
	}
	defer rows.Close()
	var reminders []Reminder
	for rows.Next() {
		var r Reminder
		var created string
		if err := rows.Scan(&r.ID, &r.Message, &created); err != nil {
			return nil, fmt.Errorf("ler lembrete: %w", err)
		}
		r.CreatedAt, _ = time.Parse(time.RFC3339Nano, created)
		reminders = append(reminders, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("listar lembretes: %w", err)
	}
	return reminders, nil
}

func (s *Store) Done(ctx context.Context, path string, position int) (bool, error) {
	if position < 1 {
		return false, nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("iniciar conclusão: %w", err)
	}
	defer tx.Rollback()
	var id int64
	err = tx.QueryRowContext(ctx,
		`SELECT id FROM reminders WHERE path = ? AND completed_at IS NULL ORDER BY id LIMIT 1 OFFSET ?`,
		path, position-1).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("localizar lembrete: %w", err)
	}
	res, err := tx.ExecContext(ctx,
		`UPDATE reminders SET completed_at = ? WHERE id = ? AND completed_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return false, fmt.Errorf("concluir lembrete: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("confirmar conclusão: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("confirmar conclusão: %w", err)
	}
	return n == 1, nil
}

func (s *Store) Clear(ctx context.Context, path string) (int64, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE reminders SET completed_at = ? WHERE path = ? AND completed_at IS NULL`,
		time.Now().UTC().Format(time.RFC3339Nano), path)
	if err != nil {
		return 0, fmt.Errorf("limpar lembretes: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("contar lembretes removidos: %w", err)
	}
	return n, nil
}

func ParsePosition(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 {
		return 0, errors.New("use um número de lembrete válido")
	}
	return n, nil
}
