// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

func TestResolveUsername(t *testing.T) {
	s := testStore(t)
	calls := 0
	s.history.api = tg.NewClient(telegram.InvokeFunc(func(_ context.Context, in bin.Encoder, out bin.Decoder) error {
		req, ok := in.(*tg.ContactsResolveUsernameRequest)
		if !ok {
			return nil
		}
		calls++
		if req.Username != "gochannel" {
			return &tgerr.Error{Code: 400, Type: "USERNAME_NOT_OCCUPIED"}
		}
		*out.(*tg.ContactsResolvedPeer) = tg.ContactsResolvedPeer{
			Peer:  &tg.PeerChannel{ChannelID: 9},
			Chats: []tg.ChatClass{&tg.Channel{ID: 9, AccessHash: 3, Title: "Go", Username: "gochannel", Broadcast: true}},
		}
		return nil
	}))
	ctx := context.Background()
	chat, err := s.ResolveUsername(ctx, "gochannel")
	if err != nil || chat.ID != peerID(&tg.PeerChannel{ChannelID: 9}) {
		t.Fatalf("chat %+v, %v", chat, err)
	}
	// Known now: no second request.
	if _, err := s.ResolveUsername(ctx, "GoChannel"); err != nil || calls != 1 {
		t.Fatalf("%v, %d calls", err, calls)
	}
	if _, err := s.ResolveUsername(ctx, "nobody"); !errors.Is(err, model.ErrLinkNotFound) {
		t.Fatalf("unknown name: %v", err)
	}
}

// A mention without a username links to its user by ID.
func TestMentionNameLinksTheUser(t *testing.T) {
	s := testStore(t)
	msgs, err := s.ingest(context.Background(), []tg.MessageClass{&tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 5}, Message: "hi Pavel", Date: 10,
		Entities: []tg.MessageEntityClass{&tg.MessageEntityMentionName{Offset: 3, Length: 5, UserID: 42}}}}, false, 0)
	if err != nil || len(msgs) != 1 {
		t.Fatal(err)
	}
	runs := model.TextRuns(msgs[0].Text, msgs[0].Entities)
	if len(runs) != 2 || runs[1].URL != "tg://user?id=42" {
		t.Fatalf("runs %+v", runs)
	}
}
