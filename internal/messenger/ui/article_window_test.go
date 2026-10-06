// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"image"
	"testing"
	"time"

	"gio-mw/defaults"
	"gio-mw/defaults/schemes"
	"gio-mw/wdk"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// wholeArticleStore gives the whole article once it is told to.
type wholeArticleStore struct {
	benchmarkHistory
	whole model.RichPage
	ready chan struct{}
}

func (s wholeArticleStore) RichMessage(ctx context.Context, _ model.MessageKey) (model.RichPage, error) {
	select {
	case <-s.ready:
		return s.whole, nil
	case <-ctx.Done():
		return model.RichPage{}, ctx.Err()
	}
}

// articleViewHarness lays out an article window's view without the window.
type articleViewHarness struct {
	t      *testing.T
	view   *articleWindow
	store  wholeArticleStore
	router input.Router
	now    time.Time
}

// newArticleViewHarness opens the part of page, whose whole is whole, at
// fragment.
func newArticleViewHarness(t *testing.T, part, whole model.RichPage, fragment string) *articleViewHarness {
	t.Helper()
	h := &articleViewHarness{t: t, store: wholeArticleStore{whole: whole, ready: make(chan struct{})}, now: time.Unix(1_790_000_000, 0)}
	h.view = newArticleView(h.store, localization.For("en"), richMessage(part), fragment, func() {})
	t.Cleanup(h.view.Close)
	h.frame()
	return h
}

func (h *articleViewHarness) frame() {
	gtx := layout.Context{Ops: new(op.Ops), Source: h.router.Source(), Now: h.now, Constraints: layout.Exact(image.Pt(500, 400)), Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Values: map[string]any{}}
	wdk.InitMaterialThemeInContext(gtx, defaults.NewTheme(gtx, schemes.SchemeBaselineLight()))
	h.view.layout(gtx, false)
	h.router.Frame(gtx.Ops)
	h.now = h.now.Add(16 * time.Millisecond)
}

// load lets the store give the whole article, and lays out the frames that
// show it.
func (h *articleViewHarness) load() {
	close(h.store.ready)
	deadline := time.Now().Add(5 * time.Second)
	for !h.view.whole && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
		h.frame()
	}
	for range 4 {
		h.frame()
	}
}

// A window opened at an anchor the part does not have shows the part and
// waits; once the whole article comes, it opens the details that hide the
// anchor and scrolls to it.
func TestArticleWindowGoesToFragmentOfWhole(t *testing.T) {
	part, whole := anchorPage(true), anchorPage(false)
	part.Blocks = part.Blocks[:10]
	h := newArticleViewHarness(t, part, whole, "deep end")
	for range 4 {
		h.frame()
	}
	if got := h.view.page.toast.Text(); got != "" {
		t.Fatalf("the part tells %q", got)
	}
	if h.view.list.Position.Offset != 0 {
		t.Fatalf("the part scrolled to %d", h.view.list.Position.Offset)
	}
	h.load()
	r := h.view.page.rows[h.view.message.Key.MessageID]
	top, ok := r.articleState.tops["deep end"]
	if !ok {
		t.Fatal("the details over the anchor stay closed")
	}
	if got := h.view.list.Position.Offset; got != articleWindowMargin+top {
		t.Fatalf("the window is at %d, the anchor at %d", got, articleWindowMargin+top)
	}
}

// An anchor the whole article does not have either is told of.
func TestArticleWindowMissingFragment(t *testing.T) {
	h := newArticleViewHarness(t, anchorPage(true), anchorPage(false), "nowhere")
	h.load()
	if got := h.view.page.toast.Text(); got != localization.For("en").T("rich.anchor_missing") {
		t.Fatalf("toast %q", got)
	}
}
