package quotahistory

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreReturnsReadingsSinceAcrossDays(t *testing.T) {
	store := NewStore(t.TempDir())
	day1 := time.Date(2026, 11, 15, 23, 50, 0, 0, time.UTC)
	day2 := day1.Add(20 * time.Minute)
	if errAppend := store.Append(
		Reading{At: day1, Provider: "claude", AuthID: "a"},
		Reading{At: day2, Provider: "claude", AuthID: "a"},
	); errAppend != nil {
		t.Fatal(errAppend)
	}

	readings, errSince := store.Since(day1.Add(time.Minute))
	if errSince != nil {
		t.Fatal(errSince)
	}
	if len(readings) != 1 || !readings[0].At.Equal(day2) {
		t.Fatalf("readings = %+v, want only the one at %s", readings, day2)
	}
}

func TestStoreSkipsTruncatedLines(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	at := time.Date(2026, 11, 16, 8, 0, 0, 0, time.UTC)
	if errAppend := store.Append(Reading{At: at, AuthID: "a"}); errAppend != nil {
		t.Fatal(errAppend)
	}
	if errWrite := appendFile(filepath.Join(dir, "2026-11-16.jsonl"), []byte(`{"at":"2026-11-16T08:05`)); errWrite != nil {
		t.Fatal(errWrite)
	}

	readings, errSince := store.Since(at.Add(-time.Hour))
	if errSince != nil {
		t.Fatal(errSince)
	}
	if len(readings) != 1 {
		t.Fatalf("got %d readings, want 1", len(readings))
	}
}

func TestStorePruneDeletesOnlyDaysBeforeCutoff(t *testing.T) {
	dir := t.TempDir()
	store := NewStore(dir)
	old := time.Date(2026, 10, 1, 12, 0, 0, 0, time.UTC)
	recent := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	if errAppend := store.Append(Reading{At: old}, Reading{At: recent}); errAppend != nil {
		t.Fatal(errAppend)
	}

	if errPrune := store.Prune(time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC)); errPrune != nil {
		t.Fatal(errPrune)
	}

	if _, errStat := os.Stat(filepath.Join(dir, "2026-10-01.jsonl")); !os.IsNotExist(errStat) {
		t.Fatalf("old day file still exists: %v", errStat)
	}
	if _, errStat := os.Stat(filepath.Join(dir, "2026-11-16.jsonl")); errStat != nil {
		t.Fatalf("recent day file missing: %v", errStat)
	}
}

func TestStoreSinceWithoutDirectoryIsEmpty(t *testing.T) {
	readings, errSince := NewStore(filepath.Join(t.TempDir(), "missing")).Since(time.Time{})
	if errSince != nil || len(readings) != 0 {
		t.Fatalf("readings = %v, err = %v", readings, errSince)
	}
}
