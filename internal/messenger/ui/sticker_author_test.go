// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"komarugram/internal/messenger/model"
	"testing"
)

func TestStickerSetAuthorCopiesID(t *testing.T) {
	h := newComposerHarness(t)
	openStickerSet(h, model.StickerSet{Title: "Cats", AuthorID: 123456789, Count: 1, Items: []model.PickerItem{{ID: "one", Emoji: "🐈"}}})
	h.click(495, 290)
	for range 30 {
		h.frame()
	}
	h.click(390, 390)
	h.frame()
	if _, data, ok := h.router.WriteClipboard(); !ok || string(data) != "123456789" {
		t.Fatalf("author ID not copied: %q %t", data, ok)
	}
	if h.p.stickers.modal.toast.Text() == "" {
		t.Fatal("no copy feedback")
	}
}

func TestStickerSetAuthorOpensChat(t *testing.T) {
	h := newComposerHarness(t)
	openStickerSet(h, model.StickerSet{Title: "Cats", AuthorID: 123, Count: 1, Items: []model.PickerItem{{Emoji: "🐈"}}})
	var opened model.Chat
	h.p.openChat = func(chat model.Chat, _ model.MessageID) { opened = chat }
	h.p.stickers.authorResult = &stickerSetResult{authorID: 123, author: &model.Chat{ID: 123, Title: "Creator"}}
	h.frame()
	if opened.ID != 123 || !h.p.stickers.modal.closing {
		t.Fatal("resolved creator did not open")
	}
}
