// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"fmt"
	"gioui.org/font"
	"gioui.org/font/gofont"
	"gioui.org/op/paint"
	"gioui.org/text"
	"image"
	"image/color"
	"image/png"
	"komarugram/internal/messenger/styledtext"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gioui.org/layout"
	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

func TestRenderTextBlocks(t *testing.T) {
	dir := os.Getenv("TEXT_BLOCKS_PNG_DIR")
	if dir == "" {
		t.Skip("set TEXT_BLOCKS_PNG_DIR to a directory")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	checkTextAfterBitmap(t, filepath.Join(dir, "text-bitmap-tail.png"))
	text, entities := mockstore.TextBlocksExample()
	for _, dark := range []bool{false, true} {
		for _, width := range []int{360, 640} {
			for _, expanded := range []bool{false, true} {
				p := newChatPage(benchmarkHistory{}, func() {})
				p.images = &imageOps{}
				m := model.Message{Key: model.MessageKey{MessageID: 1}, Text: text, Entities: entities, Date: time.Date(2026, 10, 4, 12, 30, 0, 0, time.UTC), ContentRevision: 1}
				r := &messageRow{revision: 1, runs: model.TextRuns(text, entities)}
				r.prepareTextBlocks()
				for i := range r.textBlocks {
					r.textBlocks[i].expanded = expanded
				}
				p.rows = map[model.MessageID]*messageRow{1: r}
				name := fmt.Sprintf("text-blocks-%d-dark-%t-expanded-%t.png", width, dark, expanded)
				renderToast(t, filepath.Join(dir, name), image.Pt(width, 920), dark, func(gtx layout.Context) {
					p.images.BeginFrame()
					layout.UniformInset(12).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						gtx.Constraints.Min = image.Point{}
						return p.row(gtx, m, false, 0, localization.For("ru"), false)
					})
					p.images.EndFrame()
				})
				p.Close()
			}
		}
	}
}

// A transparent bitmap changes Gio's paint material just like a color emoji.
// The vector glyph in the next 32-glyph batch must restore the text color.
type textTestEmoji struct{}

func (textTestEmoji) Match(r []rune) (int, int) {
	if len(r) > 0 && r[0] == '👋' {
		return 0, 1
	}
	return 0, 0
}
func (textTestEmoji) Image(_, size int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, size, size))
}
func checkTextAfterBitmap(t *testing.T, path string) {
	t.Helper()
	shaper := text.NewShaper(text.NoSystemFonts(), text.WithCollection(gofont.Collection()), text.WithEmojiImages(textTestEmoji{}))
	var last image.Rectangle
	renderToast(t, path, image.Pt(800, 80), false, func(gtx layout.Context) {
		paint.Fill(gtx.Ops, color.NRGBA{R: 255, G: 255, B: 255, A: 255})
		gtx.Constraints.Min = image.Point{}
		style := styledtext.Text(shaper, styledtext.SpanStyle{Font: font.Font{Typeface: "Go Mono"}, Content: strings.Repeat("a", 30) + "👋aX", Size: 20, Color: color.NRGBA{A: 255}})
		style.Decorate = func(_ layout.Context, f styledtext.Fragment, draw func()) {
			for _, c := range f.Clusters {
				if c.Start == 32 {
					last = c.Bounds
				}
			}
			draw()
		}
		style.Layout(gtx, nil)
	})
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	im, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if last.Empty() {
		t.Fatal("missing trailing glyph hit region")
	}
	for y := last.Min.Y; y < last.Max.Y; y++ {
		for x := last.Min.X; x < last.Max.X; x++ {
			r, g, b, _ := im.At(x, y).RGBA()
			if r < 0x8000 && g < 0x8000 && b < 0x8000 {
				return
			}
		}
	}
	t.Fatal("bitmap hid the following vector glyph")
}
