// SPDX-License-Identifier: Unlicense OR MIT

package tgstore

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"komarugram/internal/messenger/model"
)

func TestTextBlockMetadataSurvivesConversionCacheAndRevision(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	msg := &tg.Message{ID: 1, PeerID: &tg.PeerUser{UserID: 4}, Message: "codequote", Entities: []tg.MessageEntityClass{
		&tg.MessageEntityPre{Length: 4, Language: "go"},
		&tg.MessageEntityBlockquote{Offset: 4, Length: 5, Collapsed: true},
	}}
	m, _ := convertMessage("a", msg, nil)
	if m.Entities[0].Language != "go" || !m.Entities[1].Collapsed {
		t.Fatalf("metadata lost: %+v", m.Entities)
	}
	if err := s.Cache().SaveMessages(ctx, []model.Message{m}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Cache().Around(ctx, 4, 0, 10)
	if err != nil || len(got) != 1 {
		t.Fatalf("cache: %v, %v", got, err)
	}
	if got[0].Entities[0].Language != "go" || !got[0].Entities[1].Collapsed {
		t.Fatal("cache lost metadata")
	}
	before := model.Revision(m)
	m.Entities[0].Language = "python"
	if model.Revision(m) == before {
		t.Fatal("language edit does not invalidate layout")
	}
	before = model.Revision(m)
	m.Entities[1].Collapsed = false
	if model.Revision(m) == before {
		t.Fatal("collapse edit does not invalidate layout")
	}
}
