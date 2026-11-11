package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	_ "time/tzdata"

	_ "modernc.org/sqlite"
)

//go:embed all:web/dist
var assets embed.FS

const maxEvents = 500

type Event struct {
	ID                string `json:"id"`
	Title             string `json:"title"`
	Description       string `json:"description"`
	Category          string `json:"category"`
	Kind              string `json:"kind"`
	Theme             string `json:"theme"`
	Style             string `json:"style"`
	Holiday           string `json:"holiday,omitempty"`
	TargetAt          string `json:"target_at"`
	Timezone          string `json:"timezone"`
	Recurrence        string `json:"recurrence"`
	DurationSeconds   int64  `json:"duration_seconds"`
	PausedRemaining   *int64 `json:"paused_remaining,omitempty"`
	Pinned            bool   `json:"pinned"`
	Demo              bool   `json:"demo"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
	EffectiveTargetAt string `json:"effective_target_at,omitempty"`
}

type server struct {
	db  *sql.DB
	now func() time.Time
}

func openStore(path string, demo bool) (*server, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	fail := func(e error) (*server, error) { db.Close(); return nil, e }
	if _, err = db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;
CREATE TABLE IF NOT EXISTS events (id TEXT PRIMARY KEY, payload TEXT NOT NULL CHECK(json_valid(payload)), created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL);`); err != nil {
		return fail(err)
	}
	s := &server{db: db, now: time.Now}
	tx, err := db.Begin()
	if err != nil {
		return fail(err)
	}
	defer tx.Rollback()
	var initialized string
	err = tx.QueryRow(`SELECT value FROM settings WHERE key='initialized'`).Scan(&initialized)
	if errors.Is(err, sql.ErrNoRows) {
		if demo {
			for _, e := range examples(s.now()) {
				if err := writeEvent(tx, e); err != nil {
					return fail(err)
				}
			}
		}
		if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES('initialized','1')`); err != nil {
			return fail(err)
		}
	} else if err != nil {
		return fail(err)
	}
	if err = tx.Commit(); err != nil {
		return fail(err)
	}
	return s, nil
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}
func stamp(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}

func validate(e *Event) error {
	e.Title = strings.TrimSpace(e.Title)
	e.Category = strings.TrimSpace(e.Category)
	if len([]rune(e.Title)) < 1 || len([]rune(e.Title)) > 60 {
		return errors.New("标题需要 1–60 个字")
	}
	if len([]rune(e.Description)) > 500 {
		return errors.New("备注最多 500 个字")
	}
	if e.Category == "" {
		e.Category = "生活"
	}
	if len([]rune(e.Category)) > 16 {
		return errors.New("分类最多 16 个字")
	}
	if !oneOf(e.Kind, "countdown", "anniversary", "timer", "festival") {
		return errors.New("不支持的计时类型")
	}
	if e.Kind == "festival" {
		if !oneOf(e.Holiday, "spring", "christmas", "midautumn", "dragonboat", "newyear", "halloween", "oktoberfest", "thanksgiving", "carnival", "dayofdead", "sanfermin", "gion", "lantern", "qingming", "qixi", "double9", "laba", "eve", "national", "labor") {
			return errors.New("请选择支持的节日主题")
		}
		e.Category = "节日"
		if oneOf(e.Holiday, "spring", "midautumn", "dragonboat", "oktoberfest", "thanksgiving", "carnival", "lantern", "qingming", "qixi", "double9", "laba", "eve") && e.Recurrence != "none" {
			return errors.New("浮动日期节日请按实际日期保存，每年重新选择预设更新")
		}
	} else {
		e.Holiday = ""
	}
	if !oneOf(e.Theme, "violet", "rose", "ocean", "amber", "mint") {
		return errors.New("不支持的配色")
	}
	if !oneOf(e.Style, "flip", "digits", "ring", "minimal", "neon", "gauge", "matrix", "steps", "ticket", "analog", "hourglass", "orbit", "terminal", "note", "liquid", "calendar", "odometer", "cassette", "polaroid", "binary", "radar", "scroll", "prism", "constellation") {
		return errors.New("不支持的显示样式")
	}
	if !oneOf(e.Recurrence, "none", "daily", "weekly", "yearly") {
		return errors.New("不支持的重复方式")
	}
	if e.Timezone == "" {
		e.Timezone = "Asia/Shanghai"
	}
	if _, err := time.LoadLocation(e.Timezone); err != nil {
		return errors.New("无效时区")
	}
	if _, err := time.Parse(time.RFC3339Nano, e.TargetAt); err != nil {
		return errors.New("无效的目标日期")
	}
	if e.Kind == "timer" {
		if e.DurationSeconds < 1 || e.DurationSeconds > 604800 {
			return errors.New("短时计时需要在 1 秒到 7 天之间")
		}
		if e.Recurrence != "none" {
			return errors.New("短时计时不能设置日期重复")
		}
		if e.PausedRemaining != nil && (*e.PausedRemaining < 0 || *e.PausedRemaining > e.DurationSeconds) {
			return errors.New("无效的暂停时长")
		}
	} else {
		e.DurationSeconds = 0
		e.PausedRemaining = nil
	}
	return nil
}

// Repeating dates preserve local wall-clock time. Feb 29 falls on Feb 28 in non-leap years.
func effectiveTarget(e Event, now time.Time) string {
	target, _ := time.Parse(time.RFC3339Nano, e.TargetAt)
	if e.Recurrence == "none" || target.After(now) {
		return stamp(target)
	}
	loc, _ := time.LoadLocation(e.Timezone)
	if loc == nil {
		loc = time.UTC
	}
	target = target.In(loc)
	local := now.In(loc)
	h, m, sec := target.Clock()
	switch e.Recurrence {
	case "daily":
		target = time.Date(local.Year(), local.Month(), local.Day(), h, m, sec, 0, loc)
		if !target.After(now) {
			target = target.AddDate(0, 0, 1)
		}
	case "weekly":
		days := (int(target.Weekday()) - int(local.Weekday()) + 7) % 7
		target = time.Date(local.Year(), local.Month(), local.Day()+days, h, m, sec, 0, loc)
		if !target.After(now) {
			target = target.AddDate(0, 0, 7)
		}
	case "yearly":
		month, day := target.Month(), target.Day()
		for year := local.Year(); ; year++ {
			last := time.Date(year, month+1, 0, 0, 0, 0, 0, loc).Day()
			d := day
			if d > last {
				d = last
			}
			target = time.Date(year, month, d, h, m, sec, 0, loc)
			if target.After(now) {
				break
			}
		}
	}
	return stamp(target)
}

func examples(now time.Time) []Event {
	data := []Event{
		{Title: "下一站，去看海", Description: "把日常调成飞行模式，去收集一点蓝色。", Category: "生活", Kind: "countdown", Theme: "ocean", Style: "flip", TargetAt: stamp(now.Add(12*24*time.Hour + 8*time.Hour)), Pinned: true},
		{Title: "认真喜欢你的第 N 天", Description: "平凡的日子，也因为你有了意义。", Category: "纪念日", Kind: "anniversary", Theme: "rose", Style: "digits", TargetAt: stamp(now.AddDate(-2, -3, -12))},
		{Title: "留给自己的 25 分钟", Description: "放下手机，把注意力还给眼前的事。", Category: "专注", Kind: "timer", Theme: "mint", Style: "ring", TargetAt: stamp(now.Add(25 * time.Minute)), DurationSeconds: 1500},
		{Title: "把愿望留给生日", Description: "又长大一岁，也要继续可爱。", Category: "纪念日", Kind: "countdown", Theme: "violet", Style: "ring", TargetAt: stamp(now.AddDate(0, 1, 8)), Recurrence: "yearly"},
		{Title: "新的作品，准备登场", Description: "把脑海里的好点子，变成真的。", Category: "工作", Kind: "countdown", Theme: "amber", Style: "minimal", TargetAt: stamp(now.AddDate(0, 0, 6))},
		{Title: "下一次日落之前", Description: "今天也记得给生活留一点空白。", Category: "生活", Kind: "countdown", Theme: "violet", Style: "digits", TargetAt: stamp(now.Add(7 * time.Hour)), Recurrence: "daily"},
	}
	for i := range data {
		e := &data[i]
		e.ID = newID()
		e.Demo = true
		e.Timezone = "Asia/Shanghai"
		if e.Recurrence == "" {
			e.Recurrence = "none"
		}
		e.CreatedAt = stamp(now)
		e.UpdatedAt = stamp(now)
	}
	return data
}

type executor interface {
	Exec(string, ...any) (sql.Result, error)
}

func writeEvent(db executor, e Event) error {
	e.EffectiveTargetAt = ""
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO events(id,payload,created_at) VALUES(?,?,?) ON CONFLICT(id) DO UPDATE SET payload=excluded.payload`, e.ID, string(data), e.CreatedAt)
	return err
}
func readEvent(tx *sql.Tx, id string) (Event, error) {
	var e Event
	var data string
	err := tx.QueryRow(`SELECT payload FROM events WHERE id=?`, id).Scan(&data)
	if err != nil {
		return e, err
	}
	err = json.Unmarshal([]byte(data), &e)
	return e, err
}
func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}
func problem(w http.ResponseWriter, status int, err error) {
	respond(w, status, map[string]string{"error": err.Error()})
}
func decode(w http.ResponseWriter, r *http.Request, v any, limit int64) error {
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return errors.New("请使用 application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return errors.New("请求格式错误或内容过大")
	}
	if err := d.Decode(&struct{}{}); err != io.EOF {
		return errors.New("只允许一个 JSON 对象")
	}
	return nil
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `SELECT payload FROM events ORDER BY created_at DESC,id`)
	if err != nil {
		problem(w, 500, errors.New("读取数据失败"))
		return
	}
	defer rows.Close()
	items := []Event{}
	for rows.Next() {
		var data string
		var e Event
		if err = rows.Scan(&data); err == nil {
			err = json.Unmarshal([]byte(data), &e)
		}
		if err != nil {
			problem(w, 500, errors.New("读取数据失败"))
			return
		}
		e.EffectiveTargetAt = effectiveTarget(e, s.now())
		items = append(items, e)
	}
	if rows.Err() != nil {
		problem(w, 500, errors.New("读取数据失败"))
		return
	}
	respond(w, 200, map[string]any{"events": items, "server_time": stamp(s.now())})
}

func (s *server) save(w http.ResponseWriter, r *http.Request) {
	var in Event
	if err := decode(w, r, &in, 16<<10); err != nil {
		problem(w, 400, err)
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, errors.New("数据库繁忙"))
		return
	}
	defer tx.Rollback()
	now := s.now()
	status := 201
	if r.Method == http.MethodPut {
		old, err := readEvent(tx, r.PathValue("id"))
		if errors.Is(err, sql.ErrNoRows) {
			problem(w, 404, errors.New("倒计时不存在"))
			return
		}
		if err != nil {
			problem(w, 500, errors.New("读取数据失败"))
			return
		}
		in.ID = old.ID
		in.CreatedAt = old.CreatedAt
		in.Demo = false
		status = 200
		// Editing a running timer preserves its deadline unless its duration changes.
		if in.Kind == "timer" && old.Kind == "timer" && in.DurationSeconds == old.DurationSeconds {
			in.TargetAt = old.TargetAt
			in.PausedRemaining = old.PausedRemaining
		}
		if in.Kind == "timer" && (old.Kind != "timer" || in.DurationSeconds != old.DurationSeconds) {
			in.TargetAt = stamp(now.Add(time.Duration(in.DurationSeconds) * time.Second))
			in.PausedRemaining = nil
		}
	} else {
		var count int
		if err = tx.QueryRow(`SELECT count(*) FROM events`).Scan(&count); err != nil {
			problem(w, 500, errors.New("读取数据失败"))
			return
		}
		if count >= maxEvents {
			problem(w, 400, errors.New("最多保存 500 个倒计时"))
			return
		}
		in.ID = newID()
		in.CreatedAt = stamp(now)
		in.Demo = false
		in.PausedRemaining = nil
		if in.Kind == "timer" {
			in.TargetAt = stamp(now.Add(time.Duration(in.DurationSeconds) * time.Second))
		}
	}
	in.UpdatedAt = stamp(now)
	in.EffectiveTargetAt = ""
	if err := validate(&in); err != nil {
		problem(w, 400, err)
		return
	}
	if err = writeEvent(tx, in); err == nil {
		err = tx.Commit()
	}
	if err != nil {
		problem(w, 500, errors.New("保存数据失败"))
		return
	}
	in.EffectiveTargetAt = effectiveTarget(in, now)
	respond(w, status, in)
}

func (s *server) action(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Action string `json:"action"`
	}
	if err := decode(w, r, &in, 1024); err != nil {
		problem(w, 400, err)
		return
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, errors.New("数据库繁忙"))
		return
	}
	defer tx.Rollback()
	e, err := readEvent(tx, r.PathValue("id"))
	if errors.Is(err, sql.ErrNoRows) {
		problem(w, 404, errors.New("倒计时不存在"))
		return
	}
	if err != nil {
		problem(w, 500, errors.New("读取数据失败"))
		return
	}
	now := s.now()
	if in.Action == "pin" {
		e.Pinned = !e.Pinned
	} else if e.Kind != "timer" {
		problem(w, 400, errors.New("只有短时计时支持暂停与重置"))
		return
	} else {
		switch in.Action {
		case "pause":
			if e.PausedRemaining == nil {
				target, _ := time.Parse(time.RFC3339Nano, e.TargetAt)
				remaining := int64((target.Sub(now) + time.Second - 1) / time.Second)
				if remaining < 0 {
					remaining = 0
				}
				e.PausedRemaining = &remaining
			}
		case "resume":
			if e.PausedRemaining != nil {
				e.TargetAt = stamp(now.Add(time.Duration(*e.PausedRemaining) * time.Second))
				e.PausedRemaining = nil
			}
		case "restart":
			e.TargetAt = stamp(now.Add(time.Duration(e.DurationSeconds) * time.Second))
			e.PausedRemaining = nil
		default:
			problem(w, 400, errors.New("不支持的操作"))
			return
		}
	}
	e.Demo = false
	e.UpdatedAt = stamp(now)
	if err = writeEvent(tx, e); err == nil {
		err = tx.Commit()
	}
	if err != nil {
		problem(w, 500, errors.New("保存数据失败"))
		return
	}
	e.EffectiveTargetAt = effectiveTarget(e, now)
	respond(w, 200, e)
}

func (s *server) remove(w http.ResponseWriter, r *http.Request) {
	res, err := s.db.ExecContext(r.Context(), `DELETE FROM events WHERE id=?`, r.PathValue("id"))
	if err != nil {
		problem(w, 500, errors.New("删除失败"))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		problem(w, 404, errors.New("倒计时不存在"))
		return
	}
	respond(w, 204, nil)
}
func (s *server) clearExamples(w http.ResponseWriter, r *http.Request) {
	_, err := s.db.ExecContext(r.Context(), `DELETE FROM events WHERE json_extract(payload,'$.demo')=1`)
	if err != nil {
		problem(w, 500, errors.New("删除失败"))
		return
	}
	respond(w, 204, nil)
}
func (s *server) importEvents(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Version    int     `json:"version"`
		ExportedAt string  `json:"exported_at"`
		Events     []Event `json:"events"`
	}
	if err := decode(w, r, &in, 2<<20); err != nil {
		problem(w, 400, err)
		return
	}
	if in.Version != 1 || len(in.Events) == 0 || len(in.Events) > maxEvents {
		problem(w, 400, errors.New("备份格式错误或数量超限"))
		return
	}
	now := s.now()
	for i := range in.Events {
		e := &in.Events[i]
		if err := validate(e); err != nil {
			problem(w, 400, fmt.Errorf("第 %d 项：%w", i+1, err))
			return
		}
		e.ID = newID()
		e.Demo = false
		e.CreatedAt = stamp(now)
		e.UpdatedAt = stamp(now)
		e.EffectiveTargetAt = ""
	}
	tx, err := s.db.BeginTx(r.Context(), nil)
	if err != nil {
		problem(w, 500, errors.New("数据库繁忙"))
		return
	}
	defer tx.Rollback()
	var count int
	if err = tx.QueryRow(`SELECT count(*) FROM events`).Scan(&count); err != nil {
		problem(w, 500, errors.New("读取数据失败"))
		return
	}
	if count+len(in.Events) > maxEvents {
		problem(w, 400, errors.New("导入后将超过 500 项上限"))
		return
	}
	for _, e := range in.Events {
		if err = writeEvent(tx, e); err != nil {
			problem(w, 500, errors.New("导入失败"))
			return
		}
	}
	if err = tx.Commit(); err != nil {
		problem(w, 500, errors.New("导入失败"))
		return
	}
	respond(w, 201, map[string]int{"imported": len(in.Events)})
}

func (s *server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := s.db.PingContext(r.Context()); err != nil {
			problem(w, 503, errors.New("数据库不可用"))
			return
		}
		respond(w, 200, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /api/events", s.list)
	mux.HandleFunc("POST /api/events", s.save)
	mux.HandleFunc("PUT /api/events/{id}", s.save)
	mux.HandleFunc("DELETE /api/events/{id}", s.remove)
	mux.HandleFunc("POST /api/events/{id}/action", s.action)
	mux.HandleFunc("POST /api/events/import", s.importEvents)
	mux.HandleFunc("DELETE /api/examples", s.clearExamples)
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) { problem(w, 404, errors.New("接口不存在")) })
	web, _ := fs.Sub(assets, "web/dist")
	mux.Handle("/", http.FileServer(http.FS(web)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'")
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			if origin := r.Header.Get("Origin"); origin != "" {
				u, err := url.Parse(origin)
				if err != nil || u.Host != r.Host {
					problem(w, 403, errors.New("不允许跨站修改数据"))
					return
				}
			}
		}
		mux.ServeHTTP(w, r)
	})
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "HTTP listen address")
	path := flag.String("db", "data/countdown.db", "SQLite database path")
	demo := flag.Bool("demo", true, "seed examples on first launch")
	flag.Parse()
	s, err := openStore(*path, *demo)
	if err != nil {
		log.Fatal(err)
	}
	defer s.db.Close()
	if _, err := assets.ReadFile("web/dist/index.html"); err != nil {
		log.Fatal("前端尚未构建，请先执行 cd web && npm ci && npm run build")
	}
	httpServer := &http.Server{Addr: *addr, Handler: s.handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdown)
	}()
	log.Printf("时光舱 · http://%s · SQLite: %s", *addr, *path)
	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
