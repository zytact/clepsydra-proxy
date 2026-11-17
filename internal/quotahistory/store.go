package quotahistory

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MoeForge/moego/json"
)

const (
	dirName   = "quota-history"
	dayLayout = "2006-01-02"
	fileExt   = ".jsonl"
	Retention = 21 * 24 * time.Hour
)

type Window struct {
	ID            string     `json:"id"`
	UsedPercent   float64    `json:"used_percent"`
	ResetsAt      *time.Time `json:"resets_at,omitempty"`
	WindowSeconds int64      `json:"window_seconds,omitempty"`
	Scope         string     `json:"scope"`
}
type Reading struct {
	At        time.Time `json:"at"`
	Provider  string    `json:"provider"`
	AuthID    string    `json:"auth_id"`
	AuthIndex string    `json:"auth_index"`
	Email     string    `json:"email,omitempty"`
	Windows   []Window  `json:"windows,omitempty"`
	Error     string    `json:"error,omitempty"`
}

func Dir(configFilePath string) string {
	if strings.TrimSpace(configFilePath) == "" {
		return ""
	}
	return filepath.Join(filepath.Dir(configFilePath), dirName)
}

type Store struct {
	dir string
	mu  sync.Mutex
}

func NewStore(dir string) *Store {
	return &Store{dir: dir}
}
func (s *Store) Append(readings ...Reading) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if errMkdir := os.MkdirAll(s.dir, 0o700); errMkdir != nil {
		return fmt.Errorf("create quota history dir: %w", errMkdir)
	}
	byDay := map[string]*bytes.Buffer{}
	for _, reading := range readings {
		line, errMarshal := json.Marshal(reading)
		if errMarshal != nil {
			return fmt.Errorf("encode quota reading: %w", errMarshal)
		}
		day := reading.At.UTC().Format(dayLayout)
		if byDay[day] == nil {
			byDay[day] = &bytes.Buffer{}
		}
		byDay[day].Write(line)
		byDay[day].WriteByte('\n')
	}
	for day, buf := range byDay {
		if errWrite := appendFile(filepath.Join(s.dir, day+fileExt), buf.Bytes()); errWrite != nil {
			return errWrite
		}
	}
	return nil
}

func appendFile(path string, data []byte) (err error) {
	file, errOpen := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if errOpen != nil {
		return fmt.Errorf("open quota history file: %w", errOpen)
	}
	defer func() {
		if errClose := file.Close(); errClose != nil && err == nil {
			err = fmt.Errorf("close quota history file: %w", errClose)
		}
	}()
	if _, errWrite := file.Write(data); errWrite != nil {
		return fmt.Errorf("write quota history file: %w", errWrite)
	}
	return nil
}
func (s *Store) Since(since time.Time) ([]Reading, error) {
	days, errDays := s.days()
	if errDays != nil {
		return nil, errDays
	}
	first := since.UTC().Format(dayLayout)
	readings := []Reading{}
	for _, day := range days {
		if day < first {
			continue
		}
		data, errRead := os.ReadFile(filepath.Join(s.dir, day+fileExt))
		if errRead != nil {
			if os.IsNotExist(errRead) {
				continue
			}
			return nil, fmt.Errorf("read quota history file: %w", errRead)
		}
		scanner := bufio.NewScanner(bytes.NewReader(data))
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			var reading Reading
			if json.Unmarshal(scanner.Bytes(), &reading) != nil || reading.At.Before(since) {
				continue
			}
			readings = append(readings, reading)
		}
	}
	sort.SliceStable(readings, func(i, j int) bool { return readings[i].At.Before(readings[j].At) })
	return readings, nil
}
func (s *Store) Prune(cutoff time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	days, errDays := s.days()
	if errDays != nil {
		return errDays
	}
	last := cutoff.UTC().Format(dayLayout)
	for _, day := range days {
		if day >= last {
			continue
		}
		if errRemove := os.Remove(filepath.Join(s.dir, day+fileExt)); errRemove != nil && !os.IsNotExist(errRemove) {
			return fmt.Errorf("remove quota history file: %w", errRemove)
		}
	}
	return nil
}

func (s *Store) days() ([]string, error) {
	entries, errRead := os.ReadDir(s.dir)
	if errRead != nil {
		if os.IsNotExist(errRead) {
			return nil, nil
		}
		return nil, fmt.Errorf("list quota history dir: %w", errRead)
	}
	days := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, fileExt) {
			continue
		}
		day := strings.TrimSuffix(name, fileExt)
		if _, errParse := time.Parse(dayLayout, day); errParse != nil {
			continue
		}
		days = append(days, day)
	}
	sort.Strings(days)
	return days, nil
}
