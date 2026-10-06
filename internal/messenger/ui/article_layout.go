// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"
	"image/color"

	"gio-mw/token"
	"gio-mw/wdk"

	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// articleState is what a row keeps of its article between frames.
type articleState struct {
	// toggled are the details opened or closed against how they came, by
	// block id; toggles, their headers.
	toggled map[int]bool
	toggles map[int]*surface
	// slides are the items slideshows show; arrows, their buttons.
	slides map[int]int
	arrows map[int]*[2]surface
	// media are the rows of the article's photos and videos, by their id;
	// files, the cards of its audio and files.
	media map[string]*messageRow
	files map[string]*surface
	// buttons are the buttons of button rows, and cards those of cards, by
	// block id.
	buttons map[int][]surface
	cards   map[int][]surface
}

func (s *articleState) toggle(id int) *surface {
	if s.toggles == nil {
		s.toggles = map[int]*surface{}
	}
	t := s.toggles[id]
	if t == nil {
		t = new(surface)
		s.toggles[id] = t
	}
	return t
}

func (s *articleState) surfaces(of *map[int][]surface, id, n int) []surface {
	if *of == nil {
		*of = map[int][]surface{}
	}
	if len((*of)[id]) < n {
		(*of)[id] = append((*of)[id], make([]surface, n-len((*of)[id]))...)
	}
	return (*of)[id]
}

// articleDraw draws an article of a row, block by block.
type articleDraw struct {
	p       *chatPage
	r       *messageRow
	m       model.Message
	l       localization.Catalog
	doc     *articleDoc
	animate bool
	// textWidth is how wide the text drawn is, for an article that is as
	// wide as its text.
	textWidth int
}

// articleLayout draws the article of message m in its bubble: its blocks
// one under another, their text one area that selects and takes clicks as
// a message's text does, and what takes clicks of its own over it.
func (p *chatPage) articleLayout(gtx layout.Context, r *messageRow, m model.Message, l localization.Catalog, animate bool) layout.Dimensions {
	doc := r.article
	for i := range doc.leaves {
		p.textBlockAction(gtx, r, &doc.leaves[i], l)
	}
	p.textEvents(gtx, r, animate)
	r.text.fragments = r.text.fragments[:0]
	gtx.Constraints.Min = image.Point{}
	a := &articleDraw{p: p, r: r, m: m, l: l, doc: doc, animate: animate}
	// Controls are recorded with the blocks and laid out after the text's
	// area, so that a press on one starts no selection.
	macro := op.Record(gtx.Ops)
	size := a.stack(gtx, doc.blocks, image.Point{})
	call := macro.Stop()
	if !doc.wide {
		size.X = min(a.textWidth, size.X)
	}
	r.text.size = size
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pointer.CursorText.Add(gtx.Ops)
	r.text.clicker.Add(gtx.Ops)
	r.text.dragger.Add(gtx.Ops)
	area.Pop()
	call.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

// stack draws blocks one under another, as wide as gtx allows; origin is
// where they are in the article's text area. It returns their size.
func (a *articleDraw) stack(gtx layout.Context, blocks []*articleBlock, origin image.Point) image.Point {
	y := 0
	for i, b := range blocks {
		if i > 0 && b.kind != articleAnchor {
			y += gtx.Dp(unit.Dp(b.skip))
		}
		pos := image.Pt(0, y)
		stack := op.Offset(pos).Push(gtx.Ops)
		y += a.block(gtx, b, origin.Add(pos))
		stack.Pop()
	}
	return image.Pt(gtx.Constraints.Max.X, y)
}

// block draws b, and returns its height.
func (a *articleDraw) block(gtx layout.Context, b *articleBlock, origin image.Point) int {
	gtx.Constraints.Min = image.Point{}
	switch b.kind {
	case articleFlow:
		return a.flow(gtx, b.leaf, origin).Size.Y
	case articleCode:
		return a.p.textBlock(gtx, a.r, &a.doc.leaves[b.leaf], origin, a.l, a.animate).Size.Y
	case articleQuote:
		return a.quote(gtx, b, origin)
	case articlePullquote:
		return a.pullquote(gtx, b, origin)
	case articleList:
		return a.list(gtx, b, origin)
	case articleDivider:
		h := gtx.Dp(9)
		offset(gtx, image.Pt(0, gtx.Dp(4)), func(gtx layout.Context) layout.Dimensions {
			fillRect(gtx, scheme(gtx).OutlineVariant, image.Pt(gtx.Constraints.Max.X, max(gtx.Dp(1), 1)))
			return layout.Dimensions{}
		})
		return h
	case articleTable:
		return a.table(gtx, b, origin)
	case articleDetails:
		return a.details(gtx, b, origin)
	case articleMedia:
		return a.mediaBlock(gtx, b, origin)
	case articleButtons:
		return a.buttonRow(gtx, b)
	case articleMath:
		return a.math(gtx, b, origin)
	case articleEmbedPost:
		return a.embedPost(gtx, b, origin)
	case articleCard:
		return a.card(gtx, b, origin)
	}
	return 0
}

// flow draws the text of leaf, -1 for none.
func (a *articleDraw) flow(gtx layout.Context, leaf int, origin image.Point) layout.Dimensions {
	if leaf < 0 {
		return layout.Dimensions{}
	}
	gtx.Constraints.Min = image.Point{}
	dims := a.p.textFlow(gtx, a.r, &a.doc.leaves[leaf], origin, a.animate)
	a.textWidth = max(a.textWidth, origin.X+dims.Size.X)
	return dims
}

// below draws leaf under what is y high, skip apart, and returns the new
// height.
func (a *articleDraw) below(gtx layout.Context, leaf int, origin image.Point, x, y int, skip unit.Dp) int {
	if leaf < 0 {
		return y
	}
	if y > 0 {
		y += gtx.Dp(skip)
	}
	pos := image.Pt(x, y)
	dims := offset(gtx, pos, func(gtx layout.Context) layout.Dimensions { return a.flow(gtx, leaf, origin.Add(pos)) })
	return y + dims.Size.Y
}

// plate draws content on the plate of a quote or a code block, padded,
// with the bar of a quote when bar is set, and returns its height.
func (a *articleDraw) plate(gtx layout.Context, bar bool, content func(gtx layout.Context, pad image.Point) int) int {
	sc := scheme(gtx)
	pad := image.Pt(min(gtx.Dp(10), gtx.Constraints.Max.X/2), gtx.Dp(6))
	inner := gtx
	inner.Constraints.Max.X = max(1, gtx.Constraints.Max.X-2*pad.X)
	macro := op.Record(gtx.Ops)
	h := content(inner, pad)
	call := macro.Stop()
	size := image.Pt(gtx.Constraints.Max.X, h+2*pad.Y)
	fillRounded(gtx, sc.SurfaceVariant.Color.SetOpacity(.55), size, gtx.Dp(6))
	if bar {
		paint.FillShape(gtx.Ops, sc.Primary.Color.AsNRGBA(), clip.UniformRRect(image.Rect(0, 0, min(size.X, gtx.Dp(3)), size.Y), gtx.Dp(1)).Op(gtx.Ops))
	}
	call.Add(gtx.Ops)
	return size.Y
}

// quote draws a quote: of text as a message's quote block, collapsible; of
// blocks on a plate beside a bar; with its author under it.
func (a *articleDraw) quote(gtx layout.Context, b *articleBlock, origin image.Point) int {
	if b.leaf >= 0 {
		h := a.p.textBlock(gtx, a.r, &a.doc.leaves[b.leaf], origin, a.l, a.animate).Size.Y
		return a.below(gtx, b.caption, origin, gtx.Dp(10), h, 4)
	}
	return a.plate(gtx, true, func(gtx layout.Context, pad image.Point) int {
		var h int
		offset(gtx, pad, func(gtx layout.Context) layout.Dimensions {
			h = a.stack(gtx, b.children, origin.Add(pad)).Y
			return layout.Dimensions{}
		})
		h = a.below(gtx, b.caption, origin, pad.X, pad.Y+h, 4) - pad.Y
		return h
	})
}

// pullquote draws a quote set off in the middle, between space above and
// below, with its author under it.
func (a *articleDraw) pullquote(gtx layout.Context, b *articleBlock, origin image.Point) int {
	space := gtx.Dp(8)
	y := a.below(gtx, b.leaf, origin, 0, space, 0)
	if len(b.children) > 0 {
		pos := image.Pt(0, y)
		y += offset(gtx, pos, func(gtx layout.Context) layout.Dimensions {
			return layout.Dimensions{Size: a.stack(gtx, b.children, origin.Add(pos))}
		}).Size.Y
	}
	return a.below(gtx, b.caption, origin, 0, y, 4) + space
}

// list draws a list: each item's marker, and its text or blocks beside it.
func (a *articleDraw) list(gtx layout.Context, b *articleBlock, origin image.Point) int {
	sc := scheme(gtx)
	ty := wdk.GetMaterialTheme(gtx).Typescale[token.TypestyleBodyLarge]
	line := gtx.Sp(ty.LineHeight)
	marker, gap := gtx.Dp(26), gtx.Dp(6)
	x := marker + gap
	inner := gtx
	inner.Constraints.Max.X = max(1, gtx.Constraints.Max.X-x)
	y := 0
	for i, it := range b.items {
		if i > 0 {
			y += gtx.Dp(3)
		}
		pos := image.Pt(x, y)
		h := 0
		offset(inner, pos, func(gtx layout.Context) layout.Dimensions {
			if it.leaf >= 0 {
				h = a.flow(gtx, it.leaf, origin.Add(pos)).Size.Y
			} else {
				h = a.stack(gtx, it.blocks, origin.Add(pos)).Y
			}
			return layout.Dimensions{}
		})
		switch {
		case it.checkbox:
			box := gtx.Dp(16)
			at := image.Pt(marker-box-gtx.Dp(2), y+(line-box)/2)
			offset(gtx, at, func(gtx layout.Context) layout.Dimensions {
				rect := image.Rectangle{Max: image.Pt(box, box)}
				if it.checked {
					paint.FillShape(gtx.Ops, sc.Primary.Color.AsNRGBA(), clip.UniformRRect(rect, gtx.Dp(3)).Op(gtx.Ops))
					return exact(gtx, rect.Max, func(gtx layout.Context) layout.Dimensions { return iconDone(gtx, sc.Primary.OnColor) })
				}
				stroke(gtx, rect, gtx.Dp(3), max(gtx.Dp(2), 1), sc.SurfaceVariant.OnColor.AsNRGBA())
				return layout.Dimensions{Size: rect.Max}
			})
		case it.marker == "•":
			d := gtx.Dp(5)
			at := image.Pt(marker-d-gtx.Dp(6), y+(line-d)/2)
			paint.FillShape(gtx.Ops, sc.Surface.OnColor.AsNRGBA(), clip.Ellipse(image.Rectangle{Min: at, Max: at.Add(image.Pt(d, d))}).Op(gtx.Ops))
		case it.marker != "":
			macro := op.Record(gtx.Ops)
			mgtx := gtx
			mgtx.Constraints = layout.Constraints{Max: image.Pt(marker+gap, line)}
			dims := label(mgtx, it.marker, token.TypestyleBodyLarge, sc.Surface.OnColor, 1)
			call := macro.Stop()
			offset(gtx, image.Pt(max(0, marker-dims.Size.X), y), func(gtx layout.Context) layout.Dimensions {
				call.Add(gtx.Ops)
				return dims
			})
		}
		y += max(h, line)
	}
	return y
}

// math draws a formula. Until formulas are drawn, it is its source on a
// plate, as Telegram Desktop shows one it cannot draw.
func (a *articleDraw) math(gtx layout.Context, b *articleBlock, origin image.Point) int {
	return a.plate(gtx, false, func(gtx layout.Context, pad image.Point) int {
		return offset(gtx, pad, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min.X = gtx.Constraints.Max.X
			return a.flow(gtx, b.leaf, origin.Add(pad))
		}).Size.Y
	})
}

// embedPost draws a post embedded on a plate beside a bar: its author and
// date over its blocks, its caption under them.
func (a *articleDraw) embedPost(gtx layout.Context, b *articleBlock, origin image.Point) int {
	return a.plate(gtx, true, func(gtx layout.Context, pad image.Point) int {
		sc := scheme(gtx)
		y := pad.Y
		if b.author != "" || !b.date.IsZero() {
			head := offset(gtx, pad, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Baseline}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return label(gtx, b.author, token.TypestyleLabelLargeEmphasized, sc.Surface.OnColor, 1)
					}),
					layout.Rigid(layout.Spacer{Width: 8}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						if b.date.IsZero() {
							return layout.Dimensions{}
						}
						return label(gtx, dayOfMonth(a.l, b.date.Local(), gtx.Now, true), token.TypestyleLabelMedium, sc.SurfaceVariant.OnColor, 1)
					}))
			})
			y += head.Size.Y + gtx.Dp(4)
		}
		pos := image.Pt(pad.X, y)
		offset(gtx, pos, func(gtx layout.Context) layout.Dimensions {
			y += a.stack(gtx, b.children, origin.Add(pos)).Y
			return layout.Dimensions{}
		})
		return a.below(gtx, b.caption, origin, pad.X, y, 4) - pad.Y
	})
}

// stroke draws the outline of rect, rounded by radius, width wide.
func stroke(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	path := clip.UniformRRect(rect, radius).Path(gtx.Ops)
	paint.FillShape(gtx.Ops, col, clip.Stroke{Path: path, Width: float32(width)}.Op())
}
