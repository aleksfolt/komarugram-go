// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"errors"
	"strings"

	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"

	"komarugram/internal/messenger/model"
)

// ResolveUsername implements model.TelegramLinkSource: a peer known by the
// name first, then Telegram's, as Telegram Desktop.
func (s *Store) ResolveUsername(ctx context.Context, name string) (model.Chat, error) {
	c := s.history
	c.mu.Lock()
	api, id := c.api, int64(0)
	for pid, p := range c.peers {
		if p.Username != "" && strings.EqualFold(p.Username, name) {
			id = pid
			break
		}
	}
	c.mu.Unlock()
	if id != 0 {
		return s.chatFor(id, nil), nil
	}
	if api == nil {
		return model.Chat{}, errors.New("offline")
	}
	res, err := s.resolveUsername(ctx, api, name)
	if err != nil {
		return model.Chat{}, err
	}
	return s.chatFor(peerID(res.Peer), nil), nil
}

// resolveUsername asks Telegram whose name it is, and remembers them.
func (s *Store) resolveUsername(ctx context.Context, api *tg.Client, name string) (*tg.ContactsResolvedPeer, error) {
	res, err := api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{Username: name})
	if tgerr.Is(err, "USERNAME_NOT_OCCUPIED", "USERNAME_INVALID") {
		return nil, model.ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	s.rememberPeers(res.Users, res.Chats)
	return res, nil
}

// ChannelByID implements model.TelegramLinkSource for t.me/c links.
func (s *Store) ChannelByID(ctx context.Context, id int64) (model.Chat, error) {
	chat := peerID(&tg.PeerChannel{ChannelID: id})
	c := s.history
	c.mu.Lock()
	peer, known := c.peers[chat]
	api := c.api
	c.mu.Unlock()
	if known && peer.Hash != 0 {
		return s.chatFor(chat, nil), nil
	}
	if api == nil {
		return model.Chat{}, errors.New("offline")
	}
	res, err := api.ChannelsGetChannels(ctx, []tg.InputChannelClass{&tg.InputChannel{ChannelID: id}})
	if err != nil {
		return model.Chat{}, model.ErrLinkNotFound
	}
	for _, raw := range res.GetChats() {
		if ch, ok := raw.(*tg.Channel); ok && ch.ID == id && !ch.Min {
			s.rememberPeers(nil, []tg.ChatClass{ch})
			return s.chatFor(chat, nil), nil
		}
	}
	return model.Chat{}, model.ErrLinkNotFound
}

// CheckInvite implements model.TelegramLinkSource.
func (s *Store) CheckInvite(ctx context.Context, hash string) (model.ChatInvite, error) {
	api := s.api()
	if api == nil {
		return model.ChatInvite{}, errors.New("offline")
	}
	res, err := api.MessagesCheckChatInvite(ctx, hash)
	if tgerr.Is(err, "INVITE_HASH_EXPIRED", "INVITE_HASH_INVALID", "INVITE_HASH_EMPTY") {
		return model.ChatInvite{}, model.ErrLinkNotFound
	}
	if err != nil {
		return model.ChatInvite{}, err
	}
	switch r := res.(type) {
	case *tg.ChatInviteAlready:
		return s.inviteChat(r.Chat)
	case *tg.ChatInvitePeek:
		return s.inviteChat(r.Chat)
	case *tg.ChatInvite:
		return model.ChatInvite{Title: r.Title, Members: r.ParticipantsCount, Channel: r.Broadcast, Request: r.RequestNeeded, Paid: r.SubscriptionFormID != 0}, nil
	}
	return model.ChatInvite{}, model.ErrLinkNotFound
}

func (s *Store) inviteChat(raw tg.ChatClass) (model.ChatInvite, error) {
	s.rememberPeers(nil, []tg.ChatClass{raw})
	var p tg.PeerClass
	switch ch := raw.(type) {
	case *tg.Channel:
		p = &tg.PeerChannel{ChannelID: ch.ID}
	case *tg.Chat:
		p = &tg.PeerChat{ChatID: ch.ID}
	default:
		return model.ChatInvite{}, model.ErrLinkNotFound
	}
	chat := s.chatFor(peerID(p), nil)
	return model.ChatInvite{Chat: &chat, Title: chat.Title}, nil
}

// JoinInvite implements model.TelegramLinkSource.
func (s *Store) JoinInvite(ctx context.Context, hash string) (model.Chat, error) {
	api := s.api()
	if api == nil {
		return model.Chat{}, errors.New("offline")
	}
	res, err := api.MessagesImportChatInvite(ctx, hash)
	if tgerr.Is(err, "INVITE_REQUEST_SENT") {
		return model.Chat{}, model.ErrJoinRequested
	}
	if tgerr.Is(err, "INVITE_HASH_EXPIRED", "INVITE_HASH_INVALID") {
		return model.Chat{}, model.ErrLinkNotFound
	}
	if err != nil {
		return model.Chat{}, err
	}
	joined, ok := res.(*tg.MessagesChatInviteJoinResultOk)
	if !ok {
		return model.Chat{}, model.ErrJoinVerification
	}
	_ = s.Handle(ctx, joined.Updates)
	var chats []tg.ChatClass
	switch u := joined.Updates.(type) {
	case *tg.Updates:
		chats = u.Chats
	case *tg.UpdatesCombined:
		chats = u.Chats
	}
	for _, raw := range chats {
		if inv, err := s.inviteChat(raw); err == nil {
			return *inv.Chat, nil
		}
	}
	return model.Chat{}, model.ErrLinkNotFound
}

// UserByID implements model.TelegramLinkSource for mentions without a
// username: their messages bring the user.
func (s *Store) UserByID(_ context.Context, id int64) (model.Chat, error) {
	chat := peerID(&tg.PeerUser{UserID: id})
	s.history.mu.Lock()
	peer, ok := s.history.peers[chat]
	s.history.mu.Unlock()
	if !ok || peer.Hash == 0 && !peer.Rights.Self {
		return model.Chat{}, model.ErrLinkNotFound
	}
	return s.chatFor(chat, nil), nil
}

func (s *Store) api() *tg.Client {
	s.history.mu.Lock()
	defer s.history.mu.Unlock()
	return s.history.api
}
