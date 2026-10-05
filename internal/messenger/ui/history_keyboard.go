// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"image"

	"gio-mw/token"
	wdk "gio-mw/wdk"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"

	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
)

// The buttons under a bot's message, as Telegram Desktop's msgBotKbButton.
const (
	inlineKeyHeight  = unit.Dp(36)
	inlineKeyGap     = unit.Dp(2)
	inlineKeyPadding = unit.Dp(10)
	inlineKeyIcon    = unit.Dp(12)
)

// withKeyboard draws a bubble as wide as m's buttons need and the buttons
// under it.
func (p *chatPage) withKeyboard(gtx layout.Context, r *messageRow, m model.Message, join bubbleJoin, l localization.Catalog, bubble layout.Widget) layout.Dimensions {
	gtx.Constraints.Min.X = min(p.inlineKeysWidth(gtx, m, l), gtx.Constraints.Max.X)
	dims := bubble(gtx)
	top := dims.Size.Y + gtx.Dp(inlineKeyGap)
	kb := offset(gtx, image.Pt(0, top), func(gtx layout.Context) layout.Dimensions {
		return p.inlineKeys(gtx, r, m, dims.Size.X, shapeOf(gtx, join, m.Outgoing), l)
	})
	return layout.Dimensions{Size: image.Pt(dims.Size.X, top+kb.Size.Y)}
}

func (p *chatPage) inlineKeysWidth(gtx layout.Context, m model.Message, l localization.Catalog) int {
	gtx.Constraints.Min = image.Point{}
	pad, gap := gtx.Dp(inlineKeyPadding), gtx.Dp(inlineKeyGap)
	widest := 0
	for _, row := range m.Buttons {
		w := 0
		for _, b := range row {
			macro := op.Record(gtx.Ops)
			dims := label(gtx, inlineKeyText(b, l), token.TypestyleLabelLarge, token.MatColor{}, 1)
			macro.Stop()
			w += dims.Size.X + 2*pad + gap
		}
		widest = max(widest, w-gap)
	}
	return widest
}

func inlineKeyText(b model.MessageButton, l localization.Catalog) string {
	switch b.Kind {
	case "url", "webview", "simple_webview", "callback", "copy":
		return b.Text
	}
	return b.Text + " · " + l.T("history.readonly")
}

func (p *chatPage) inlineKeys(gtx layout.Context, r *messageRow, m model.Message, width int, outer bubbleShape, l localization.Catalog) layout.Dimensions {
	height, gap, small := gtx.Dp(inlineKeyHeight), gtx.Dp(inlineKeyGap), gtx.Dp(bubbleRadiusJoined)
	y := 0
	for i, row := range m.Buttons {
		x := 0
		for j, b := range row {
			w := (width - gap*(len(row)-1)) / len(row)
			if j == len(row)-1 {
				w = width - x
			}
			shape := bubbleShape{small, small, small, small}
			if i == len(m.Buttons)-1 {
				if j == 0 {
					shape.sw = outer.sw
				}
				if j == len(row)-1 {
					shape.se = outer.se
				}
			}
			offset(gtx, image.Pt(x, y), func(gtx layout.Context) layout.Dimensions {
				return p.inlineKey(gtx, r, m, i, j, b, image.Pt(w, height), shape, l)
			})
			x += w + gap
		}
		y += height + gap
	}
	return layout.Dimensions{Size: image.Pt(width, max(0, y-gap))}
}

func (p *chatPage) inlineKey(gtx layout.Context, r *messageRow, m model.Message, y, x int, b model.MessageButton, size image.Point, shape bubbleShape, l localization.Catalog) layout.Dimensions {
	s := &r.buttons[y][x]
	if s.Clicked(gtx) {
		switch b.Kind {
		case "url":
			p.askLink(b.URL)
		case "webview", "simple_webview":
			p.pressWebView(gtx, p.chat, m, b)
		case "callback", "copy":
			p.pressButton(gtx, m, y, x, b, l)
		}
	}
	sc := scheme(gtx)
	fg := sc.SecondaryContainer.OnColor
	text := inlineKeyText(b, l)
	if b.Kind == "callback" && p.pending(m.Key, y, x) {
		text += " …"
	}
	var icon wdk.IconWidget
	switch b.Kind {
	case "url":
		icon = iconOpenInNew
	case "webview", "simple_webview":
		icon = iconOpenInBrowser
	case "copy":
		icon = iconCopy
	}
	defer shape.rrect(size).Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, sc.SecondaryContainer.Color.AsNRGBA())
	style := surfaceStyle{background: fg.SetOpacity(0), content: fg, button: b.Text}
	return s.Layout(gtx, size, style, func(gtx layout.Context) layout.Dimensions {
		pad := gtx.Dp(inlineKeyPadding)
		gtx.Constraints = layout.Exact(image.Pt(max(0, size.X-2*pad), size.Y))
		offset(gtx, image.Pt(pad, 0), func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				return label(gtx, text, token.TypestyleLabelLarge, fg, 1)
			})
		})
		if icon != nil {
			d, inset := gtx.Dp(inlineKeyIcon), gtx.Dp(4)
			offset(gtx, image.Pt(size.X-d-inset, inset), func(gtx layout.Context) layout.Dimensions {
				return exact(gtx, image.Pt(d, d), func(gtx layout.Context) layout.Dimensions { return icon(gtx, fg.SetOpacity(.8)) })
			})
		}
		return layout.Dimensions{Size: size}
	})
}
