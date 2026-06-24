package measurement

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Event struct {
	TS         string `json:"ts"`
	Kind       string `json:"kind"`
	Detail     string `json:"detail"`
	BytesSaved int    `json:"bytesSaved"`
}

func eventsPath(dataHome string) string {
	return filepath.Join(dataHome, "events.jsonl")
}

func DBPath(dataHome string) string {
	return eventsPath(dataHome)
}

func Init(dataHome string) error {
	return os.MkdirAll(dataHome, 0o755)
}

func Record(dataHome, kind, detail string, bytesSaved int) error {
	if err := Init(dataHome); err != nil {
		return err
	}
	ev := Event{
		TS:         time.Now().UTC().Format(time.RFC3339),
		Kind:       kind,
		Detail:     detail,
		BytesSaved: bytesSaved,
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(eventsPath(dataHome), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(append(data, '\n'))
	return err
}

type Summary struct {
	Days         int
	TotalEvents  int
	Compressions int
	BytesSaved   int
	Sessions     int
}

func Report(dataHome string, days int) (Summary, error) {
	s := Summary{Days: days}
	if days <= 0 {
		days = 7
		s.Days = days
	}
	since := time.Now().UTC().AddDate(0, 0, -days)
	path := eventsPath(dataHome)
	f, err := os.Open(path)
	if err != nil {
		return s, nil
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var ev Event
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		ts, err := time.Parse(time.RFC3339, ev.TS)
		if err != nil || ts.Before(since) {
			continue
		}
		s.TotalEvents++
		if ev.Kind == "compress" {
			s.Compressions++
			s.BytesSaved += ev.BytesSaved
		}
		if ev.Kind == "session" {
			s.Sessions++
		}
	}
	return s, nil
}

func FormatSummary(s Summary) string {
	return fmt.Sprintf(
		"Last %d days: %d events | %d compressions | ~%d bytes saved | %d sessions logged",
		s.Days, s.TotalEvents, s.Compressions, s.BytesSaved, s.Sessions,
	)
}
