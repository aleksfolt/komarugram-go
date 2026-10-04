// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"io"
	"strings"
	"unicode"
	"unicode/utf8"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/styledtext"

	"gio-mw/token"
	"gioui.org/io/clipboard"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// messageTextBlock is a flow of inline runs, optionally wrapped in a code or
// quote plate. All hit regions retain offsets into the original message.
type messageTextBlock struct {
	first, end, runeStart int
	trimStart, trimEnd    bool
	clusters              []styledtext.Cluster
	action                surface
	expanded              bool
}

func (r *messageRow) prepareTextBlocks() {
	if r.textBlocks != nil {
		return
	}
	runeStart := 0
	for i := 0; i < len(r.runs); {
		end, n := i+1, utf8.RuneCountInString(r.runs[i].Text)
		for end < len(r.runs) && r.runs[end].Block == r.runs[i].Block {
			n += utf8.RuneCountInString(r.runs[end].Text)
			end++
		}
		b := messageTextBlock{first: i, end: end, runeStart: runeStart}
		if r.runs[i].Block == 0 {
			// The block itself supplies a line break. Keep the source newline
			// for copying, but do not draw a second empty line beside it.
			b.trimStart = i > 0 && strings.HasPrefix(r.runs[i].Text, "\n")
			b.trimEnd = end < len(r.runs) && strings.HasSuffix(r.runs[end-1].Text, "\n")
		}
		r.textBlocks = append(r.textBlocks, b)
		runeStart += n
		i = end
	}
}

func (r *messageRow) copyTextBlock(b *messageTextBlock) string {
	if r.noCopy {
		return ""
	}
	var text strings.Builder
	for _, run := range r.runs[b.first:b.end] {
		if run.Spoiler && !r.revealed {
			text.WriteString("[•••]")
		} else {
			text.WriteString(run.Text)
		}
	}
	return text.String()
}

func (p *chatPage) richText(gtx layout.Context, r *messageRow, l localization.Catalog, animate bool) layout.Dimensions {
	end := p.trace.Begin("history.rich-text")
	defer end()
	r.prepareTextBlocks()
	for i := range r.textBlocks {
		b := &r.textBlocks[i]
		if b.action.Clicked(gtx) {
			run := r.runs[b.first]
			if run.Pre {
				if text := r.copyTextBlock(b); text != "" {
					gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
					p.toast.Show(l.T("text.copied"))
				}
			} else if run.Quote && run.Collapsed {
				b.expanded = !b.expanded
				gtx.Execute(op.InvalidateCmd{})
			}
		}
	}
	p.textEvents(gtx, r, animate)
	r.text.fragments = r.text.fragments[:0]
	gtx.Constraints.Min = image.Point{}
	macro := op.Record(gtx.Ops)
	size := image.Point{}
	for i := range r.textBlocks {
		b := &r.textBlocks[i]
		if b.end-b.first == 1 && r.runs[b.first].Block == 0 && r.runs[b.first].Text == "\n" && (b.trimStart || b.trimEnd) {
			continue
		}
		if size.Y > 0 {
			size.Y += gtx.Dp(6)
		}
		origin := image.Pt(0, size.Y)
		stack := op.Offset(origin).Push(gtx.Ops)
		dims := p.textBlock(gtx, r, b, origin, l, animate)
		stack.Pop()
		size.X = max(size.X, dims.Size.X)
		size.Y += dims.Size.Y
	}
	call := macro.Stop()
	r.text.size = size
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pointer.CursorText.Add(gtx.Ops)
	r.text.clicker.Add(gtx.Ops)
	r.text.dragger.Add(gtx.Ops)
	area.Pop()
	// Controls are registered after the text area, so their presses cannot
	// start a text selection or activate a link underneath a button.
	call.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

func (p *chatPage) textBlock(gtx layout.Context, r *messageRow, b *messageTextBlock, origin image.Point, l localization.Catalog, animate bool) layout.Dimensions {
	run := r.runs[b.first]
	if run.Block == 0 {
		return p.textFlow(gtx, r, b, origin, animate)
	}
	pad := min(gtx.Dp(10), gtx.Constraints.Max.X/2)
	inner := gtx
	inner.Constraints.Max.X = max(1, gtx.Constraints.Max.X-2*pad)
	pos := image.Pt(pad, gtx.Dp(6))
	macro := op.Record(gtx.Ops)
	if run.Pre {
		stack := op.Offset(pos).Push(gtx.Ops)
		header := layout.Flex{Alignment: layout.Middle}.Layout(inner,
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return label(gtx, codeLanguageLabel(run.Language), token.TypestyleLabelMedium, scheme(gtx).SurfaceVariant.OnColor, 1)
			}),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if r.noCopy {
					return layout.Dimensions{}
				}
				return textButton(gtx, &b.action, l.T("text.copy_code"))
			}))
		stack.Pop()
		pos.Y += header.Size.Y + gtx.Dp(4)
	}
	stack := op.Offset(pos).Push(gtx.Ops)
	dims := p.textFlow(inner, r, b, origin.Add(pos), animate)
	stack.Pop()
	pos.Y += dims.Size.Y
	if run.Quote && run.Collapsed {
		stack := op.Offset(pos).Push(gtx.Ops)
		key := "text.expand_quote"
		if b.expanded {
			key = "text.collapse_quote"
		}
		button := textButton(inner, &b.action, l.T(key))
		stack.Pop()
		pos.Y += button.Size.Y
	}
	size := image.Pt(gtx.Constraints.Max.X, pos.Y+gtx.Dp(6))
	call := macro.Stop()
	fillRounded(gtx, scheme(gtx).SurfaceVariant.Color.SetOpacity(.55), size, gtx.Dp(6))
	if run.Quote {
		paint.FillShape(gtx.Ops, scheme(gtx).Primary.Color.AsNRGBA(), clip.UniformRRect(image.Rect(0, 0, min(size.X, gtx.Dp(3)), size.Y), gtx.Dp(1)).Op(gtx.Ops))
	}
	call.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

// Language is a label from untrusted input, never a filename or an instruction.
// Bound it before shaping and omit control/bidi characters in the header.
func codeLanguageLabel(language string) string {
	var out strings.Builder
	count := 0
	for _, r := range language {
		if count == 40 {
			break
		}
		count++
		if !unicode.IsControl(r) && !unicode.Is(unicode.Cf, r) {
			out.WriteRune(r)
		}
	}
	return strings.TrimSpace(out.String())
}
