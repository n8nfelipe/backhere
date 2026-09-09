package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "backhere.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestReminderLifecycleAndDirectoryIsolation(t *testing.T) {
	ctx := context.Background()
	s := newTestStore(t)
	if err := s.Add(ctx, "/one", "first"); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(ctx, "/one", "second"); err != nil {
		t.Fatal(err)
	}
	if err := s.Add(ctx, "/two", "other"); err != nil {
		t.Fatal(err)
	}

	one, err := s.Active(ctx, "/one")
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 2 || one[0].Message != "first" || one[1].Message != "second" {
		t.Fatalf("unexpected reminders: %#v", one)
	}
	two, err := s.Active(ctx, "/two")
	if err != nil {
		t.Fatal(err)
	}
	if len(two) != 1 || two[0].Message != "other" {
		t.Fatalf("unexpected isolation: %#v", two)
	}

	ok, err := s.Done(ctx, "/one", 1)
	if err != nil || !ok {
		t.Fatalf("Done = %v, %v", ok, err)
	}
	one, _ = s.Active(ctx, "/one")
	if len(one) != 1 || one[0].Message != "second" {
		t.Fatalf("after done: %#v", one)
	}

	n, err := s.Clear(ctx, "/one")
	if err != nil || n != 1 {
		t.Fatalf("Clear = %d, %v", n, err)
	}
	one, _ = s.Active(ctx, "/one")
	if len(one) != 0 {
		t.Fatalf("after clear: %#v", one)
	}
	if untouched, _ := s.Active(ctx, "/two"); len(untouched) != 1 {
		t.Fatalf("other directory changed: %#v", untouched)
	}
}

func TestDoneMissing(t *testing.T) {
	ok, err := newTestStore(t).Done(context.Background(), "/none", 1)
	if err != nil || ok {
		t.Fatalf("Done = %v, %v", ok, err)
	}
}

func BenchmarkActiveNoReminders(b *testing.B) {
	s, err := Open(filepath.Join(b.TempDir(), "backhere.db"))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.Active(ctx, "/empty"); err != nil {
			b.Fatal(err)
		}
	}
}
