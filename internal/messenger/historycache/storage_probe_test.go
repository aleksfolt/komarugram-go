package historycache

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/security"
)

// TestStorageProbe measures a synthetic cache, plaintext and encrypted, and
// what deleting media and compacting give back. Data and Storage, phase 0:
// STORAGE_PROBE=1 runs it; STORAGE_PROBE_MEDIA and STORAGE_PROBE_MESSAGES
// change its size, STORAGE_PROBE_TEMP=file VACUUM's temporary storage.
func TestStorageProbe(t *testing.T) {
	if os.Getenv("STORAGE_PROBE") == "" {
		t.Skip("STORAGE_PROBE not set")
	}
	for _, encrypted := range []bool{false, true} {
		name := "plain"
		if encrypted {
			name = "adiantum"
		}
		t.Run(name, func(t *testing.T) { probeStorage(t, encrypted) })
	}
}

func probeStorage(t *testing.T, encrypted bool) {
	ctx := context.Background()
	dir := t.TempDir()
	var p *security.Manager
	if encrypted {
		var e error
		if p, e = security.OpenPath(filepath.Join(dir, "security"), &fakeTPM{}); e != nil {
			t.Fatal(e)
		}
		if e = p.Enable(ctx, "probe"); e != nil {
			t.Fatal(e)
		}
	}
	path := filepath.Join(dir, "history")
	c, e := Open(path, "a", p)
	if e != nil {
		t.Fatal(e)
	}
	defer c.Close()
	file := path + ".plain"
	if encrypted {
		file = path + ".secure"
	}
	report := func(stage string) {
		t.Logf("%-26s %s", stage, probeFile(t, c, file))
	}

	messages, media := envInt("STORAGE_PROBE_MESSAGES", 100000), envInt("STORAGE_PROBE_MEDIA", 400)
	const chats, mediaSize = 1000, 512 << 10
	rnd := rand.New(rand.NewSource(1))
	start := time.Now()
	batch := make([]model.Message, 0, 500)
	for i := range messages {
		batch = append(batch, model.Message{
			Key:  model.MessageKey{AccountID: "a", ChatID: int64(i%chats + 1), MessageID: model.MessageID(i/chats + 1)},
			Text: probeText(rnd),
		})
		if len(batch) == cap(batch) || i == messages-1 {
			if e = c.SaveMessages(ctx, batch); e != nil {
				t.Fatal(e)
			}
			batch = batch[:0]
		}
	}
	t.Logf("%d messages in %d chats: %v", messages, chats, time.Since(start).Round(time.Millisecond))
	for {
		done, e := c.indexBatch()
		if e != nil {
			t.Fatal(e)
		}
		if done {
			break
		}
	}
	report("messages + FTS")

	blob := make([]byte, mediaSize)
	var slowest time.Duration
	start = time.Now()
	for i := range media {
		rnd.Read(blob)
		s := time.Now()
		if e = c.SaveMedia(ctx, fmt.Sprintf("media/%d", i), blob); e != nil {
			t.Fatal(e)
		}
		slowest = max(slowest, time.Since(s))
	}
	t.Logf("%d media of %d KiB: %v, slowest SaveMedia %v", media, mediaSize>>10, time.Since(start).Round(time.Millisecond), slowest.Round(time.Microsecond))
	report("media saved")

	for _, q := range []string{
		`SELECT count(*), coalesce(sum(length(data)),0) FROM media`,
		`SELECT count(*), coalesce(sum(length(payload)),0) FROM messages`,
		`SELECT count(*), coalesce(sum(n),0) FROM (SELECT chat, count(*) n FROM messages GROUP BY chat)`,
	} {
		var n, bytes int64
		s := time.Now()
		if e = c.db.QueryRowContext(ctx, q).Scan(&n, &bytes); e != nil {
			t.Fatal(e)
		}
		t.Logf("%v  %s → %d, %d", time.Since(s).Round(time.Microsecond), q, n, bytes)
	}

	if _, e = c.db.ExecContext(ctx, `DELETE FROM media WHERE CAST(substr(key,7) AS INTEGER)%2=0`); e != nil {
		t.Fatal(e)
	}
	report("half the media deleted")
	if _, e = c.db.ExecContext(ctx, `PRAGMA incremental_vacuum`); e != nil {
		t.Fatal(e)
	}
	report("incremental_vacuum")

	if v := os.Getenv("STORAGE_PROBE_TEMP"); v != "" {
		if _, e = c.db.ExecContext(ctx, `PRAGMA temp_store=`+v); e != nil {
			t.Fatal(e)
		}
	}
	s := time.Now()
	if _, e = c.db.ExecContext(ctx, `PRAGMA auto_vacuum=INCREMENTAL; VACUUM`); e != nil {
		t.Fatal(e)
	}
	t.Logf("VACUUM (to auto_vacuum=INCREMENTAL): %v", time.Since(s).Round(time.Millisecond))
	report("after VACUUM")

	if _, e = c.db.ExecContext(ctx, `DELETE FROM media WHERE CAST(substr(key,7) AS INTEGER)%4=1`); e != nil {
		t.Fatal(e)
	}
	report("another quarter deleted")
	s = time.Now()
	if _, e = c.db.ExecContext(ctx, `PRAGMA incremental_vacuum(4096)`); e != nil {
		t.Fatal(e)
	}
	t.Logf("incremental_vacuum(4096): %v", time.Since(s).Round(time.Millisecond))
	report("incremental_vacuum(4096)")
	s = time.Now()
	if _, e = c.db.ExecContext(ctx, `PRAGMA incremental_vacuum`); e != nil {
		t.Fatal(e)
	}
	t.Logf("incremental_vacuum: %v", time.Since(s).Round(time.Millisecond))
	report("incremental_vacuum")
	c.Close()
	runtime.GC()
	t.Logf("%-26s %s", "closed", probeRSS())
}

func probeFile(t *testing.T, c *Cache, file string) string {
	var pageSize, pages, free, auto int64
	var journal string
	for q, v := range map[string]any{
		`PRAGMA page_size`: &pageSize, `PRAGMA page_count`: &pages, `PRAGMA freelist_count`: &free,
		`PRAGMA journal_mode`: &journal, `PRAGMA auto_vacuum`: &auto,
	} {
		if e := c.db.QueryRow(q).Scan(v); e != nil {
			t.Fatal(q, e)
		}
	}
	st, e := os.Stat(file)
	if e != nil {
		t.Fatal(e)
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return fmt.Sprintf("file %s, pages %d×%d, free %d (%s), journal %s, auto_vacuum %d, heap %s, %s",
		mib(st.Size()), pages, pageSize, free, mib(free*pageSize), journal, auto, mib(int64(ms.HeapInuse)), probeRSS())
}

func probeText(r *rand.Rand) string {
	words := []string{"привет", "hello", "storage", "кеш", "message", "telegram", "файл", "photo", "ok", "завтра"}
	var b strings.Builder
	for range 4 + r.Intn(30) {
		b.WriteString(words[r.Intn(len(words))])
		b.WriteByte(' ')
	}
	return b.String()
}

// probeRSS reads the resident set and its peak; Linux only.
func probeRSS() string {
	f, e := os.Open("/proc/self/status")
	if e != nil {
		return "RSS n/a"
	}
	defer f.Close()
	var out []string
	for s := bufio.NewScanner(f); s.Scan(); {
		if k, v, ok := strings.Cut(s.Text(), ":"); ok && (k == "VmRSS" || k == "VmHWM") {
			out = append(out, k+" "+strings.TrimSpace(v))
		}
	}
	return strings.Join(out, ", ")
}

func mib(n int64) string { return fmt.Sprintf("%.1f MiB", float64(n)/(1<<20)) }

func envInt(name string, def int) int {
	if n, e := strconv.Atoi(os.Getenv(name)); e == nil && n > 0 {
		return n
	}
	return def
}
