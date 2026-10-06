// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"strings"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/pointer"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

func richMessage(page model.RichPage) model.Message {
	summary := page.Summary()
	return model.Message{Key: model.MessageKey{ChatID: 100, MessageID: 1}, Text: summary.Text, Entities: summary.Entities, Rich: &page, ContentRevision: 1}
}

func richText(s string) model.RichText { return model.RichText{Text: s} }

// An article's text is its blocks' texts, each after a line break that is
// not drawn; a block that shows nothing is left out, and an article of text
// only is as wide as its text.
func TestPrepareArticle(t *testing.T) {
	page := model.RichPage{Blocks: []model.RichBlock{
		{Kind: model.RichHeading, Level: 2, Text: richText("Заголовок 👋")},
		{Kind: model.RichParagraph},
		{Kind: model.RichParagraph, Text: richText("Абзац")},
		{Kind: model.RichList, Items: []model.RichListItem{{Text: richText("пункт")}, {Checkbox: true, Checked: true, Text: richText("задача")}}},
	}}
	doc := prepareArticle(page, localization.For("ru"), time.Now())
	if len(doc.blocks) != 3 || len(doc.leaves) != 4 || doc.wide {
		t.Fatalf("%d blocks, %d leaves, wide %v", len(doc.blocks), len(doc.leaves), doc.wide)
	}
	var all strings.Builder
	for _, r := range doc.runs {
		all.WriteString(r.Text)
	}
	if got := all.String(); got != "Заголовок 👋\nАбзац\nпункт\nзадача" {
		t.Fatalf("text %q", got)
	}
	for _, leaf := range doc.leaves {
		before := 0
		for _, r := range doc.runs[:leaf.first] {
			before += len([]rune(r.Text))
		}
		if leaf.runeStart != before {
			t.Fatalf("a leaf starts at rune %d, not %d", leaf.runeStart, before)
		}
	}
	list := doc.blocks[2]
	if list.items[0].marker != "•" || !list.items[1].checkbox || !list.items[1].checked {
		t.Fatalf("items %+v", list.items)
	}
	if doc := prepareArticle(model.RichPage{Blocks: []model.RichBlock{{Kind: model.RichDivider}}}, localization.For("ru"), time.Now()); !doc.wide {
		t.Fatal("a divider does not take the width")
	}
}

// Cells spanning rows and columns take their places as in HTML.
func TestTablePlaces(t *testing.T) {
	rows := []articleRow{
		{cells: []articleCell{{colspan: 1, rowspan: 2}, {colspan: 2, rowspan: 1}}},
		{cells: []articleCell{{colspan: 1, rowspan: 1}, {colspan: 1, rowspan: 1}}},
		{cells: []articleCell{{colspan: 3, rowspan: 9}}},
	}
	places := tablePlaces(rows)
	want := [][]tablePlace{
		{{0, 0, 1, 2}, {0, 1, 2, 1}},
		{{1, 1, 1, 1}, {1, 2, 1, 1}},
		{{2, 0, 3, 1}},
	}
	for i := range want {
		for j := range want[i] {
			if places[i][j] != want[i][j] {
				t.Fatalf("cell %d,%d at %+v, want %+v", i, j, places[i][j], want[i][j])
			}
		}
	}
	if got := tableColumns(rows); got != 3 {
		t.Fatalf("%d columns", got)
	}
}

// Text selected across an article's blocks copies as their texts, one to a
// line; the markers of lists are not its text.
func TestArticleSelectsAcrossBlocks(t *testing.T) {
	page := model.RichPage{Blocks: []model.RichBlock{
		{Kind: model.RichHeading, Level: 1, Text: richText("Заголовок")},
		{Kind: model.RichList, Ordered: true, Items: []model.RichListItem{{Text: richText("первый")}, {Text: richText("второй")}}},
		{Kind: model.RichTable, Rows: []model.RichTableRow{
			{Cells: []model.RichTableCell{{Text: richText("ячейка")}, {Text: richText("ещё")}}},
			{Cells: []model.RichTableCell{{Text: richText("ниже"), Colspan: 2}}},
		}},
		{Kind: model.RichParagraph, Text: richText("конец")},
	}}
	h := newEntityHarness(t, richMessage(page), model.KindUser)
	fragments := h.row.text.fragments
	first, last := fragments[0].Bounds, fragments[len(fragments)-1].Bounds
	h.pointerDrag(f32.Pt(float32(first.Min.X), float32(first.Min.Y+2)), f32.Pt(float32(last.Max.X), float32(last.Max.Y-2)))
	if got := h.row.selectedText(); got != "Заголовок\nпервый\nвторой\nячейка\nещё\nниже\nконец" {
		t.Fatalf("selected %q", got)
	}
	// A table's cells are where their rows are drawn, under the blocks
	// before it, the second row under the first.
	at := func(text string) image.Rectangle {
		for _, f := range h.row.text.fragments {
			if h.row.runs[f.Index].Text == text {
				return f.Bounds
			}
		}
		t.Fatalf("no fragment says %q", text)
		return image.Rectangle{}
	}
	if cell, row2, item := at("ячейка"), at("ниже"), at("второй"); cell.Min.Y <= item.Max.Y || row2.Min.Y <= cell.Max.Y || at("конец").Min.Y <= row2.Max.Y {
		t.Fatalf("cells at %v and %v, after %v", cell, row2, item)
	}
}

// Details open and close by their header, and show their blocks only open.
func TestArticleDetailsToggle(t *testing.T) {
	page := model.RichPage{Blocks: []model.RichBlock{
		{Kind: model.RichDetails, Text: richText("Подробнее"), Blocks: []model.RichBlock{{Kind: model.RichParagraph, Text: richText("скрытое")}}},
	}}
	h := newEntityHarness(t, richMessage(page), model.KindUser)
	shows := func(text string) bool {
		for _, f := range h.row.text.fragments {
			if strings.Contains(h.row.runs[f.Index].Text, text) {
				return true
			}
		}
		return false
	}
	if shows("скрытое") {
		t.Fatal("closed details show their blocks")
	}
	h.clickText("Подробнее")
	h.frame()
	if !shows("скрытое") {
		t.Fatal("opened details hide their blocks")
	}
	h.clickText("Подробнее")
	h.frame()
	if shows("скрытое") {
		t.Fatal("closed again, details show their blocks")
	}
}

// A link button and a button in the text ask to open their link; a code
// block of an article copies its text.
func TestArticleButtonsAndCode(t *testing.T) {
	var label model.RichText
	label.Append(richText("Нажмите "))
	from := model.UTF16Len(label.Text)
	label.Append(richText("здесь"))
	label.Mark(from, model.Entity{Kind: "button", Button: &model.MessageButton{Kind: "url", URL: "https://example.com/text"}})
	page := model.RichPage{Blocks: []model.RichBlock{
		{Kind: model.RichButtons, Buttons: []model.RichButton{{Text: richText("Сайт"), Button: model.MessageButton{Kind: "url", URL: "https://example.com/row"}}}},
		{Kind: model.RichParagraph, Text: label},
		{Kind: model.RichCode, Language: "go", Text: richText("fmt.Println(1)")},
	}}
	h := newEntityHarness(t, richMessage(page), model.KindUser)
	h.click(image.Pt(100, 17))
	if h.page.link != "https://example.com/row" {
		t.Fatalf("the row's button asks to open %q", h.page.link)
	}
	h.page.linkModal.Close()
	h.page.link = ""
	h.frame()
	h.clickText("здесь")
	if h.page.link != "https://example.com/text" {
		t.Fatalf("the text's button asks to open %q", h.page.link)
	}
	code := &h.row.article.leaves[len(h.row.article.leaves)-1]
	code.action.click.Click()
	h.frame()
	if _, copied, ok := h.router.WriteClipboard(); !ok || string(copied) != "fmt.Println(1)" {
		t.Fatalf("copied %q, %v", copied, ok)
	}
}

// pointerDrag presses at from, drags to to and releases there.
func (h *entityHarness) pointerDrag(from, to f32.Point) {
	h.now = h.now.Add(time.Second)
	for _, e := range []pointer.Event{
		{Kind: pointer.Press, Position: from},
		{Kind: pointer.Move, Position: to},
		{Kind: pointer.Release, Position: to},
	} {
		e.Source, e.Buttons, e.Time = pointer.Mouse, pointer.ButtonPrimary, time.Duration(h.now.UnixNano())
		h.router.Queue(e)
		h.frame()
	}
}
