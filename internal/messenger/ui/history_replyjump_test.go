// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"testing"

	"komarugram/internal/messenger/model"

	"gioui.org/f32"
	"gioui.org/io/pointer"
)

// A click on the quote of a reply is seen while the list is being laid out;
// the jump must wait for the next layout, or the list is drawn from a
// position it has just left.
func TestReplyQuoteJumpWaitsForNextLayout(t *testing.T) {
	h := newChatInputHarnessOf(t, 200, func(m model.History) model.ConversationStore {
		m.Messages[149].ReplyToMessageID = 50
		return benchmarkHistory{h: m}
	})
	p := h.page
	h.seek(149, 0)
	if p.list.Position.First != 149 {
		t.Fatalf("seek stopped at %d", p.list.Position.First)
	}
	for y := float32(8); y < 120 && p.jumpPending == 0; y += 4 {
		at := f32.Pt(100+60, y)
		h.send(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Position: at, Buttons: pointer.ButtonPrimary})
		h.send(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: at})
	}
	if p.jumpPending == 0 && p.list.Position.First == 149 {
		t.Fatal("no click reached the quote")
	}
	// The frame that saw the click has not moved the list.
	if p.list.Position.First != 149 {
		t.Fatalf("the click moved the list during its layout: first %d", p.list.Position.First)
	}
	h.frame()
	if first := p.list.Position.First; first != 49 {
		t.Fatalf("the next layout stopped at %d, want 49", first)
	}
}
