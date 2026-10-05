// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"context"
	"testing"
	"time"

	"komarugram/internal/messenger/mockstore"
	"komarugram/internal/messenger/model"
)

type linkStore struct {
	*mockstore.Store
	joined bool
}

func (s *linkStore) ResolveUsername(_ context.Context, name string) (model.Chat, error) {
	if name == "durov" {
		return model.Chat{ID: 77, Title: "Durov"}, nil
	}
	return model.Chat{}, model.ErrLinkNotFound
}
func (s *linkStore) ChannelByID(context.Context, int64) (model.Chat, error) {
	return model.Chat{}, model.ErrLinkNotFound
}
func (s *linkStore) UserByID(_ context.Context, id int64) (model.Chat, error) {
	return model.Chat{ID: id}, nil
}
func (s *linkStore) CheckInvite(context.Context, string) (model.ChatInvite, error) {
	return model.ChatInvite{Title: "Go club", Members: 12}, nil
}
func (s *linkStore) JoinInvite(_ context.Context, hash string) (model.Chat, error) {
	s.joined = hash == "AbCd"
	return model.Chat{ID: 88, Title: "Go club"}, nil
}

// Links of Telegram open chats, posts and invites in the client instead of
// asking to open the browser.
func TestTelegramLinksOpenInTheClient(t *testing.T) {
	h := newComposerHarness(t)
	store := &linkStore{Store: h.p.source.(*mockstore.Store)}
	h.p.source = store
	type opened struct {
		chat int64
		post model.MessageID
	}
	var got []opened
	h.p.openChat = func(c model.Chat, post model.MessageID) { got = append(got, opened{c.ID, post}) }
	wait := func(cond func() bool) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for h.frame(); !cond(); h.frame() {
			if time.Now().After(deadline) {
				t.Fatal("timed out")
			}
			time.Sleep(5 * time.Millisecond)
		}
	}

	h.p.askLink("https://t.me/durov/5")
	wait(func() bool { return len(got) == 1 })
	if got[0] != (opened{77, 5}) || h.p.link != "" {
		t.Fatalf("opened %v, asked for the browser %q", got, h.p.link)
	}

	h.p.askLink("https://t.me/nosuchname")
	wait(func() bool { return !h.p.tgLinks.busy })
	if len(got) != 1 || h.p.link != "" {
		t.Fatalf("an unknown name opened %v", got)
	}

	h.p.askLink("https://t.me/+AbCd")
	wait(func() bool { return h.p.tgLinks.modal.Shown() })
	h.p.resolveTelegramLink(model.TelegramLink{Kind: model.LinkInvite, Hash: h.p.tgLinks.hash}, true)
	wait(func() bool { return len(got) == 2 })
	if !store.joined || got[1] != (opened{88, 0}) {
		t.Fatalf("joined %v, opened %v", store.joined, got)
	}

	h.p.askLink("tg://user?id=5")
	wait(func() bool { return len(got) == 3 })
	if got[2] != (opened{5, 0}) {
		t.Fatalf("a mention opened %v", got)
	}

	h.p.askLink("https://telegram.org")
	if h.p.link != "https://telegram.org" {
		t.Fatalf("a site did not ask for the browser: %q", h.p.link)
	}
}
