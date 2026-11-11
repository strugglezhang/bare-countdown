package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testServer(t *testing.T) (*server, *time.Time) {
	t.Helper()
	s, err := openStore(filepath.Join(t.TempDir(), "test.db"), false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.db.Close() })
	now := time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return now }
	return s, &now
}
func request(t *testing.T, s *server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var b bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&b).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	r := httptest.NewRequest(method, path, &b)
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handler().ServeHTTP(w, r)
	return w
}
func exampleInput(now time.Time) Event {
	return Event{Title: "期待周末", Category: "生活", Kind: "countdown", Style: "flip", Theme: "violet", Recurrence: "none", Timezone: "Asia/Shanghai", TargetAt: stamp(now.Add(72 * time.Hour))}
}
func create(t *testing.T, s *server, input Event) Event {
	t.Helper()
	w := request(t, s, "POST", "/api/events", input)
	if w.Code != 201 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	var e Event
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatal(err)
	}
	return e
}
func count(t *testing.T, s *server) int {
	t.Helper()
	w := request(t, s, "GET", "/api/events", nil)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	var out struct {
		Events []Event `json:"events"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	return len(out.Events)
}

func TestPersistenceAndCRUD(t *testing.T) {
	path := filepath.Join(t.TempDir(), "persistent.db")
	s, err := openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	e := create(t, s, exampleInput(time.Now()))
	e.Title = "周末 '); DROP TABLE events; --"
	w := request(t, s, "PUT", "/api/events/"+e.ID, e)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	s.db.Close()
	s, err = openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	if count(t, s) != 1 {
		t.Fatal("event lost after reopen")
	}
	w = request(t, s, "GET", "/api/events", nil)
	var out struct {
		Events []Event `json:"events"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.Events[0].Title != e.Title {
		t.Fatal("title not preserved")
	}
	if w = request(t, s, "DELETE", "/api/events/"+e.ID, nil); w.Code != 204 {
		t.Fatal(w.Body.String())
	}
	if count(t, s) != 0 {
		t.Fatal("delete did not persist")
	}
	if w = request(t, s, "DELETE", "/api/events/"+e.ID, nil); w.Code != 404 {
		t.Fatal("missing event should be 404")
	}
}

func TestTimerPauseResumeAndRestart(t *testing.T) {
	s, now := testServer(t)
	in := exampleInput(*now)
	in.Kind = "timer"
	in.DurationSeconds = 120
	e := create(t, s, in)
	*now = now.Add(31 * time.Second)
	w := request(t, s, "POST", "/api/events/"+e.ID+"/action", map[string]string{"action": "pause"})
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	json.Unmarshal(w.Body.Bytes(), &e)
	if e.PausedRemaining == nil || *e.PausedRemaining != 89 {
		t.Fatalf("paused remaining: %v", e.PausedRemaining)
	}
	*now = now.Add(2 * time.Hour)
	w = request(t, s, "POST", "/api/events/"+e.ID+"/action", map[string]string{"action": "resume"})
	e = Event{}
	json.Unmarshal(w.Body.Bytes(), &e)
	deadline, err := time.Parse(time.RFC3339Nano, e.TargetAt)
	if err != nil || deadline.Sub(*now) != 89*time.Second || e.PausedRemaining != nil {
		t.Fatalf("resume did not preserve remaining duration: %s", w.Body.String())
	}
	w = request(t, s, "POST", "/api/events/"+e.ID+"/action", map[string]string{"action": "restart"})
	json.Unmarshal(w.Body.Bytes(), &e)
	deadline, _ = time.Parse(time.RFC3339Nano, e.TargetAt)
	if deadline.Sub(*now) != 120*time.Second {
		t.Fatal("restart did not reset timer")
	}
}

func TestDateRecurrenceTimezoneAndLeapDay(t *testing.T) {
	cases := []struct{ name, target, now, zone, repeat, want string }{
		{"local daily", "2026-11-09T09:30:00+08:00", "2026-11-10T02:00:00Z", "Asia/Shanghai", "daily", "2026-11-11T01:30:00Z"},
		{"weekly", "2026-11-02T09:00:00+08:00", "2026-11-10T02:00:00Z", "Asia/Shanghai", "weekly", "2026-11-16T01:00:00Z"},
		{"leap birthday", "2024-02-29T08:00:00+08:00", "2025-01-01T00:00:00Z", "Asia/Shanghai", "yearly", "2025-02-28T00:00:00Z"},
		{"DST wall clock", "2026-03-07T09:00:00-05:00", "2026-03-08T14:00:00Z", "America/New_York", "daily", "2026-03-09T13:00:00Z"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			now, _ := time.Parse(time.RFC3339, tc.now)
			e := Event{TargetAt: tc.target, Timezone: tc.zone, Recurrence: tc.repeat}
			got, _ := time.Parse(time.RFC3339Nano, effectiveTarget(e, now))
			want, _ := time.Parse(time.RFC3339, tc.want)
			if !got.Equal(want) {
				t.Fatalf("got %s want %s", got, want)
			}
		})
	}
}

func TestImportAtomicAndPreservesDeadline(t *testing.T) {
	s, now := testServer(t)
	valid := exampleInput(*now)
	invalid := valid
	invalid.Title = ""
	w := request(t, s, "POST", "/api/events/import", map[string]any{"version": 1, "events": []Event{valid, invalid}})
	if w.Code != 400 || count(t, s) != 0 {
		t.Fatal("invalid batch partially imported")
	}
	w = request(t, s, "POST", "/api/events/import", map[string]any{"version": 1, "events": []Event{valid, valid}})
	if w.Code != 201 || count(t, s) != 2 {
		t.Fatal(w.Body.String())
	}
	w = request(t, s, "GET", "/api/events", nil)
	var out struct {
		Events []Event `json:"events"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	if out.Events[0].ID == out.Events[1].ID || out.Events[0].TargetAt != valid.TargetAt {
		t.Fatal("import corrupted identity or deadline")
	}
}

func TestValidationAndOrigin(t *testing.T) {
	s, now := testServer(t)
	for _, mutate := range []func(*Event){func(e *Event) { e.Title = " " }, func(e *Event) { e.Timezone = "invalid/zone" }, func(e *Event) { e.TargetAt = "not-a-date" }, func(e *Event) { e.Style = "bad" }, func(e *Event) { e.Kind = "timer"; e.DurationSeconds = -2 }} {
		e := exampleInput(*now)
		mutate(&e)
		if w := request(t, s, "POST", "/api/events", e); w.Code != 400 {
			t.Fatal("invalid input accepted", w.Body.String())
		}
	}
	e := exampleInput(*now)
	data, _ := json.Marshal(e)
	r := httptest.NewRequest("POST", "http://localhost:8080/api/events", bytes.NewReader(data))
	r.Header.Set("Origin", "https://evil.example")
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.handler().ServeHTTP(w, r)
	if w.Code != 403 || count(t, s) != 0 {
		t.Fatal("cross-origin write allowed")
	}
	r = httptest.NewRequest("POST", "/api/events", bytes.NewReader(data))
	w = httptest.NewRecorder()
	s.handler().ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal("non-JSON write allowed")
	}
}

func TestExamplesAreOnlySeededOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "seed.db")
	s, err := openStore(path, true)
	if err != nil {
		t.Fatal(err)
	}
	if count(t, s) != 6 {
		t.Fatal("missing examples")
	}
	w := request(t, s, "DELETE", "/api/examples", nil)
	if w.Code != 204 || count(t, s) != 0 {
		t.Fatal("clear examples failed")
	}
	s.db.Close()
	s, err = openStore(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	if count(t, s) != 0 {
		t.Fatal("deleted examples were re-seeded")
	}
}

func TestEditedExampleSurvivesClearing(t *testing.T) {
	s, err := openStore(filepath.Join(t.TempDir(), "seed.db"), true)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	w := request(t, s, http.MethodGet, "/api/events", nil)
	var out struct {
		Events []Event `json:"events"`
	}
	json.Unmarshal(w.Body.Bytes(), &out)
	e := out.Events[0]
	e.Title = "我自己的期待"
	w = request(t, s, "PUT", "/api/events/"+e.ID, e)
	if w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	request(t, s, "DELETE", "/api/examples", nil)
	if count(t, s) != 1 {
		t.Fatal("edited user event incorrectly removed")
	}
}

func TestAdditionalClockStylesSurviveRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "styles.db")
	s, err := openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, style := range []string{"neon", "gauge", "matrix", "steps", "ticket", "analog", "hourglass", "orbit", "terminal", "note", "liquid", "calendar", "odometer", "cassette", "polaroid", "binary", "radar", "scroll", "prism", "constellation"} {
		e := exampleInput(time.Now())
		e.Style = style
		e.Title = "样式 " + style
		created := create(t, s, e)
		if created.Style != style {
			t.Fatalf("style lost on save: %s", style)
		}
	}
	s.db.Close()
	s, err = openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	w := request(t, s, "GET", "/api/events", nil)
	var out struct {
		Events []Event `json:"events"`
	}
	if err = json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Events) != 20 {
		t.Fatalf("styles lost after reopen: %d", len(out.Events))
	}
	for _, e := range out.Events {
		if e.Title != "样式 "+e.Style {
			t.Fatal("saved style changed after restart")
		}
	}
}

func TestFestivalThemesPersistAndValidate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "festivals.db")
	s, err := openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := os.ReadFile("web/src/holiday-catalog.json")
	if err != nil {
		t.Fatal(err)
	}
	var extras []struct {
		Value string `json:"value"`
		Lunar []int  `json:"lunar"`
		Rule  string `json:"rule"`
		Table bool   `json:"table"`
	}
	if err = json.Unmarshal(catalog, &extras); err != nil {
		t.Fatal(err)
	}
	supported := []string{"spring", "christmas", "midautumn", "dragonboat", "newyear", "halloween"}
	for _, h := range extras {
		supported = append(supported, h.Value)
		if len(h.Lunar) > 0 || h.Rule != "" || h.Table {
			e := exampleInput(time.Now())
			e.Kind = "festival"
			e.Holiday = h.Value
			e.Recurrence = "yearly"
			if validate(&e) == nil {
				t.Fatalf("floating date holiday repeats on a fixed date: %s", h.Value)
			}
		}
	}
	for _, holiday := range supported {
		e := exampleInput(time.Now())
		e.Kind = "festival"
		e.Holiday = holiday
		e.Category = "生活"
		e.Recurrence = "none"
		saved := create(t, s, e)
		if saved.Holiday != holiday || saved.Category != "节日" {
			t.Fatalf("festival not saved: %+v", saved)
		}
	}
	s.db.Close()
	s, err = openStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	defer s.db.Close()
	if count(t, s) != len(supported) {
		t.Fatal("festival records lost on restart")
	}
	var data struct {
		Events []Event `json:"events"`
	}
	if err = json.Unmarshal(request(t, s, "GET", "/api/events", nil).Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	for _, e := range data.Events {
		if e.Holiday == "" || e.Kind != "festival" {
			t.Fatal("festival theme lost after reopen")
		}
	}
	invalid := exampleInput(time.Now())
	invalid.Kind = "festival"
	invalid.Holiday = "unknown"
	if validate(&invalid) == nil {
		t.Fatal("unknown holiday accepted")
	}
	invalid.Holiday = "spring"
	invalid.Recurrence = "yearly"
	if validate(&invalid) == nil {
		t.Fatal("lunar festival incorrectly repeats on Gregorian date")
	}
	invalid.Kind = "countdown"
	if err = validate(&invalid); err != nil || invalid.Holiday != "" {
		t.Fatal("switching kind must clear holiday decoration")
	}
}
