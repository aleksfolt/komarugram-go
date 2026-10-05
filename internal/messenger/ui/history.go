// SPDX-License-Identifier: Unlicense OR MIT

package ui

import (
	"fmt"
	"image"
	"net/url"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"komarugram/internal/diagnostics"
	"komarugram/internal/messenger/chatmedia"
	"komarugram/internal/messenger/fonts"
	"komarugram/internal/messenger/localization"
	"komarugram/internal/messenger/model"
	"komarugram/internal/messenger/preferences"
	"komarugram/internal/messenger/styledtext"
	"komarugram/pkg/player"

	"gio-mw/token"
	"gio-mw/wdk"
	"gio-mw/widget/scroll"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
)

type messageRow struct {
	// mediaTold is the failure to load the row's media told in the toast
	// last, so that a failure is told once, not on every frame.
	mediaTold   error
	avatarPoint image.Point
	bodySize    image.Point
	bodyTop     int
	album       map[model.MessageID]*messageRow
	revision    uint64
	runs        []model.TextRun
	text        textInteraction
	buttons     [][]surface
	// reactions are the message's reaction chips, which choose or take
	// back a reaction.
	reactions []surface
	// shownReactions are the message's reactions in the order of the
	// chips.
	shownReactions []model.Reaction
	// comments is the bar that opens a channel post's comments.
	comments surface
	// reply is the quote of the message replied to, which shows it.
	reply surface
	// quick takes double clicks on the bubble, which react to it.
	quick    gesture.Click
	media    widget.Clickable
	sticker  surface
	audio    audioRow
	revealed bool
	// tile shows the variant of the media chosen for tileSize pixels.
	tile     model.Message
	tileSize image.Point
}
type chatPage struct {
	membership membershipControl
	// membershipNotice reports outcomes in the chat list after the dialog is removed.
	membershipNotice func(string)
	// pinned is the bar of the chat's pinned messages.
	pinned pinnedBar
	// reacted lists who reacted to a message; edits, the versions of an
	// edited one.
	reacted reactedDialog
	edits   editsDialog
	// translation shows a message translated.
	translation translateDialog
	// chatSearch searches the chat; chatMenu is the menu of its header.
	chatSearch chatSearch
	chatMenu   chatMenu
	// highlight is the message a search went to, tinted until
	// highlightUntil; infoAsked, set when the menu asks for the chat's info.
	highlight      model.MessageID
	highlightUntil time.Time
	infoAsked      bool
	// themeAsked is set when the menu asks for the chat's theme, which
	// themeShown keeps for the info opened.
	themeAsked, themeShown bool
	// filter hides messages, as the settings ask; filtered counts what it
	// hid in the open chat, and showFiltered are the chats that show it.
	filter       *messageFilter
	filterShown  *messageFilter
	filtered     int
	showFiltered map[int64]bool
	// addFilter saves a pattern made from selected text.
	addFilter func(preferences.FilterPattern)
	deletion  messageDeletion
	stickers  stickerSetDialog
	// dialogStickers are the media of the sticker sets shown in the dialog
	// in this chat; switching chats forgets them.
	dialogStickers map[string]bool
	// releaseMemory gives memory a view dropped back to the system later;
	// keepMemory cancels that when a view is shown again.
	releaseMemory, keepMemory func()
	emojiPacks                emojiPacksDialog
	messageMenu               messageMenu
	composer                  *messageComposer
	header                    widget.Clickable
	headAvatar                widget.Clickable
	appearance                *chatThemeController
	files                     *attachmentFiles
	trace                     *diagnostics.Trace
	selection                 messageSelection
	keyboard                  struct{}
	activeText                *messageRow
	actions                   selectionBar
	forwarding                forwardPicker
	// toast tells, over the end of the history, what was done in the chat
	// and what failed there.
	toast     toast
	images    *imageOps
	revision  uint64
	dates     []string
	dayStart  []bool
	joins     []bubbleJoin
	nextDay   []int
	avatar    avatarLayout
	kind      model.ChatKind
	linkModal modal
	source    model.ConversationStore
	media     *chatmedia.Manager
	animate   bool
	chat      int64
	list      scroll.List
	messages  []model.Message
	rows      map[model.MessageID]*messageRow
	// textRunes are the rune counts of the messages' texts, which the
	// estimates of unmeasured heights need on every change of width.
	textRunes                                  map[model.MessageID]textRunes
	env                                        model.RenderEnvironment
	heights                                    *model.HeightIndex
	measures                                   map[model.MessageID]model.MessageLayout
	dirty                                      map[model.MessageID]model.MessageLayout
	restored                                   bool
	view                                       model.Viewport
	saved                                      time.Time
	newer, retry, toStart, toEnd, open, cancel surface
	link                                       string
	// jump is a jump to an end of a thread, waiting for its load.
	jump jump
	// flights are the messages that fly from the composer to their places.
	flights map[model.MessageID]*sendFlight
	// mediaError is a failure of a background task on the chat's media,
	// reported from its goroutine and told in the toast.
	errorMu    sync.Mutex
	mediaError error
	invalidate func()
	// online counts the open group's members online.
	online groupOnline
	// audio plays voice messages and music: the window's pages share one
	// player, which goes on when another chat is shown. audioBar shows
	// what it plays over the history.
	audio    *audioPlayer
	audioBar audioBar
	// openPhoto shows a photo in the viewer; nil leaves photos inline.
	openPhoto func(model.Message)
	openAlone func(model.Message)
	// openChat opens a chat, at a message unless it is 0.
	openChat  func(model.Chat, model.MessageID)
	openAudio func(model.Message)
	tgLinks   tgLinks
	// openComments shows the comments to a channel post; nil hides the
	// comments bar. thread is set for the page that shows comments.
	openComments func(model.Message)
	thread       bool
	// bot is what the page shows of bots: reply keyboards and the Start
	// button. openWebApp opens the Mini App of a bot's button; nil leaves
	// those buttons off.
	bot        botPage
	openWebApp func(gtx layout.Context, p *chatPage, req model.WebViewRequest, button string)
	// topic is set when the thread shown is a topic of a forum, whose
	// messages are read like a chat's.
	topic bool
	// threadRoot is the root of the thread shown, which replies to it do
	// not quote.
	threadRoot model.MessageID
	// classic reports whether the composer is a bar below the history
	// rather than a capsule floating over it; nil means floating.
	classic func() bool
	// overlays says how the menus and toasts are drawn; nil draws them
	// opaque. bd is the recording of the history behind the overlays in the
	// frame drawn last, nil when nothing blurs it.
	overlays func() overlayPrefs
	bd       *blurBackdrop
	// blur reports whether to blur the history behind the floating
	// composer; nil means not to.
	blur func() bool
	// frozen replaces the composer when the account is frozen.
	frozen *frozenView
	// loader shows that history is loading; fileLoader, that a file is
	// being downloaded to open.
	loader, fileLoader loadingIndicator
	// chats are the account's chats, which messages can be forwarded to.
	chats func() []model.Chat
	// title is the open chat's, for the initials of its avatar.
	title string
	// snapshotting is set while the rows of a snapshot are laid out; shot
	// is the dialog that makes one; historyWidth, how wide the history was
	// in the last frame, which a snapshot is too.
	snapshotting bool
	shot         shotDialog
	historyWidth int
	// player and setPlayer read and save the external player the user
	// chose; playerChoice asks for it.
	player      func() player.Kind
	setPlayer   func(player.Kind)
	playerPaths func() map[player.Kind]string
	// audioExternal reports whether voice messages and music open in the
	// external player, as chosen in the settings, rather than in the
	// client.
	audioExternal func() bool
	playerChoice  playerChoice
}

// classicComposer reports whether the composer is a bar below the history.
func (p *chatPage) classicComposer() bool {
	return p.composer != nil && p.classic != nil && p.classic()
}

// overlayPrefs is how the page's overlays are drawn.
func (p *chatPage) overlayPrefs() overlayPrefs {
	if p.overlays == nil {
		return overlayPrefs{opacity: float32(composerBlurOpacity)}
	}
	return p.overlays()
}

// menuBackdrop is what the page's menus blur, nil for opaque ones.
func (p *chatPage) menuBackdrop() *blurBackdrop {
	if p.overlayPrefs().menus {
		return p.bd
	}
	return nil
}

// toastBackdrop is what the page's toast blurs, nil for an opaque one.
func (p *chatPage) toastBackdrop() *blurBackdrop {
	if p.overlayPrefs().toasts {
		return p.bd
	}
	return nil
}

// blurComposer reports whether the history behind the composer is blurred.
func (p *chatPage) blurComposer() bool {
	return p.composer != nil && !p.classicComposer() && p.blur != nil && p.blur()
}

func newChatPage(source model.ConversationStore, changed func()) *chatPage {
	return &chatPage{composer: newMessageComposer(source, changed), source: source, media: chatmedia.New(source, changed), invalidate: changed, audio: newAudioPlayer()}
}

// dropStickerLoops drops, at the end of the frame, the decoded loops of the
// stickers a closed view showed, keeping their first frames for when it is
// shown again, and gives the memory back to the system a while later.
func (p *chatPage) dropStickerLoops() {
	p.media.DropLoops(func(freed int64) {
		if freed >= 1<<20 && p.releaseMemory != nil {
			p.releaseMemory()
		}
	})
}

// stickerViewShown cancels the release of the memory a sticker view
// dropped when it closed: shown again, it decodes its loops again.
func (p *chatPage) stickerViewShown() {
	if p.keepMemory != nil {
		p.keepMemory()
	}
}

// forgetDialogStickers drops the first frames of the sets shown in the
// dialog, once the chat is left: they are not likely to be shown again. The
// composer's picker is the same in every chat, and what it shows stays.
func (p *chatPage) forgetDialogStickers() {
	p.stickers.remember(p)
	var ids []string
	for id := range p.dialogStickers {
		if p.composer == nil || !p.composer.pickerShows(id) {
			ids = append(ids, id)
		}
	}
	p.media.Forget(ids)
	p.dialogStickers = nil
}

// forget drops the open chat without saving where it was, so that the
// next Layout opens it again from where the store says: see
// model.MessageRevealer.
func (p *chatPage) forget() {
	p.chat = 0
}

func (p *chatPage) Close() {
	if w, ok := p.source.(model.ChatWatcher); ok {
		w.WatchChat(p, 0)
	}
	p.stickers.stop()
	p.emojiPacks.stop()
	p.closeMenu()
	if p.composer != nil {
		p.composer.cancel()
		p.composer.forgetPasted(nil)
	}
	p.save(true)
	p.audio.stop()
	p.media.Close()
	if p.files != nil {
		p.files.Close()
	}
}

// photos returns the photos of the loaded history, album members included.
func (p *chatPage) photos() []model.Message {
	var out []model.Message
	for _, m := range p.messages {
		if len(m.Attachments) > 1 {
			out = append(out, m.Attachments...)
		} else {
			out = append(out, m)
		}
	}
	return out
}
func (p *chatPage) save(force bool) {
	if len(p.messages) == 0 || p.chat == 0 {
		return
	}
	pos := p.list.Position
	if pos.First >= len(p.messages) {
		return
	}
	v := model.Viewport{ChatID: p.chat, AnchorMessageID: p.messages[pos.First].Key.MessageID, AnchorOffsetPx: pos.Offset, AtEnd: !pos.BeforeEnd, Environment: p.env, UpdatedAt: time.Now()}
	if p.heights != nil {
		v.AbsoluteOffsetPx = p.heights.Prefix(pos.First) + int64(pos.Offset)
	}
	if !force && time.Since(p.saved) < 500*time.Millisecond {
		return
	}
	if !force && v.AnchorMessageID == p.view.AnchorMessageID && v.AnchorOffsetPx == p.view.AnchorOffsetPx && v.AtEnd == p.view.AtEnd && v.Environment == p.view.Environment && len(p.dirty) == 0 {
		return
	}
	var ls []model.MessageLayout
	for _, l := range p.dirty {
		ls = append(ls, l)
	}
	p.source.SaveView(v, ls)
	if p.trace != nil {
		p.trace.Recorder.Event(p.trace.Window, "history.save-view", fmt.Sprintf("anchor:%d offset:%d end:%t", v.AnchorMessageID, v.AnchorOffsetPx, v.AtEnd), 0, len(ls))
	}
	clear(p.dirty)
	p.view = v
	p.saved = time.Now()
}

// Layout draws the bar of what plays, if anything does, the chat's pinned
// bar, if it has pinned messages, and its history under them.
func (p *chatPage) Layout(gtx layout.Context, c model.Chat, l localization.Catalog, animate bool) layout.Dimensions {
	playing := p.audioBarSize(gtx)
	bar := playing + p.pinnedHeight(gtx, c.ID)
	if bar == 0 {
		return p.layoutHistory(gtx, c, l, animate)
	}
	size := gtx.Constraints.Max
	body := gtx
	body.Constraints = layout.Exact(image.Pt(size.X, max(size.Y-bar, 0)))
	offset(body, image.Pt(0, bar), func(gtx layout.Context) layout.Dimensions {
		return p.layoutHistory(gtx, c, l, animate)
	})
	// What plays is told over the pinned message, as in Telegram Desktop.
	offset(gtx, image.Pt(0, playing), func(gtx layout.Context) layout.Dimensions {
		p.layoutPinned(gtx, c.ID, l)
		return layout.Dimensions{}
	})
	if playing > 0 {
		p.layoutAudioBar(gtx, l, false)
	}
	return layout.Dimensions{Size: size}
}

func (p *chatPage) layoutHistory(gtx layout.Context, c model.Chat, l localization.Catalog, animate bool) layout.Dimensions {
	p.animate = animate
	p.updateDelete(c.ID, l)
	p.updateMembership(c.ID, l)
	p.trace = diagnostics.From(gtx.Values)
	if p.trace != nil {
		p.trace.History = diagnostics.History{Window: p.trace.Window, Chat: c.ID}
		end := p.trace.Begin("history.total")
		defer func() {
			end()
			h := p.trace.History
			h.Messages = len(p.messages)
			h.Visible = p.list.Position.Count
			h.First = p.list.Position.First
			h.Offset = p.list.Position.Offset
			h.RowsCached = len(p.rows)
			h.MeasurementsCached = len(p.measures)
			h.Dirty = len(p.dirty)
			h.Revision = p.revision
			if p.heights != nil {
				h.TotalHeight = p.heights.Total()
			}
			p.trace.Recorder.History(h)
			p.trace = nil
		}()
	}
	if p.chat != c.ID {
		p.forgetDialogStickers()
		p.stickers.stop()
		p.emojiPacks.stop()
		p.reacted.stop()
		p.edits.stop()
		p.shot.stop()
		p.translation.stop()
		p.closeChatSearch()
		p.chatMenu.open = false
		p.closeMenu()
		p.save(true)
		p.clearSelection()
		p.activeText = nil
		p.chat = c.ID
		p.revision = 0
		p.list = scroll.List{List: layout.List{Axis: layout.Vertical, ScrollToEnd: true}}
		p.messages = nil
		p.rows = map[model.MessageID]*messageRow{}
		p.measures = nil
		p.dirty = map[model.MessageID]model.MessageLayout{}
		p.restored = false
		p.jump = jumpNone
		p.flights = nil
		p.link = ""
		p.linkModal.Hide()
	}
	p.keyboardEvents(gtx)
	p.kind, p.title = c.Kind, c.Title
	p.menuUpdate(gtx, l)
	p.updateForward(l)
	p.source.OpenChat(c.ID)
	if w, ok := p.source.(model.ChatWatcher); ok {
		w.WatchChat(p, c.ID)
	}
	snapshotEnd := p.trace.Begin("history.snapshot+albums")
	if p.filterShown != p.filter {
		// Other filters: the history is read and filtered again.
		p.filterShown = p.filter
		p.revision = 0
	}
	var history model.History
	fresh := true
	if source, ok := p.source.(interface {
		HistorySince(int64, uint64) (model.History, bool)
	}); ok {
		history, fresh = source.HistorySince(c.ID, p.revision)
	} else {
		history = p.source.History(c.ID)
	}
	p.threadRoot = history.ThreadRoot
	if fresh {
		history.Messages = p.filterMessages(model.GroupAlbums(history.Messages), c.ID)
		p.revision = history.Revision
	} else {
		history.Messages = p.messages
	}
	snapshotEnd()
	if p.trace != nil {
		p.trace.History.SnapshotFresh = fresh
	}
	sc := scheme(gtx)
	size := gtx.Constraints.Max
	// The history is recorded to be drawn again, blurred, behind whatever
	// overlay blurs it: the composer, a menu, the toast.
	overlays := p.overlayPrefs()
	recording := p.blurComposer() || overlays.menus || overlays.toasts
	p.bd = nil
	var record op.MacroOp
	if recording {
		record = op.Record(gtx.Ops)
	}
	fillRect(gtx, sc.SurfaceContainerLow, size)
	if p.appearance != nil {
		p.appearance.Update(c.ID, token.IsDarkColorSet(sc.Surface))
		p.appearance.Background(gtx)
	}
	p.historyWidth = size.X
	theme := uint32(sc.Surface.Color.AsNRGBA().R)<<16 | uint32(sc.Surface.Color.AsNRGBA().G)<<8 | uint32(sc.Surface.Color.AsNRGBA().B)
	env := model.RenderEnvironment{WidthPx: size.X, ScaleMilli: int(gtx.Metric.PxPerDp * 1000), TextScaleMilli: int(gtx.Metric.PxPerSp * 1000), Locale: string(l.Language()), FontRevision: fonts.Revision(), ThemeRevision: theme, RendererRevision: 11}
	if p.trace != nil {
		p.trace.History.Environment = fmt.Sprintf("width:%d dp:%d sp:%d locale:%s font:%d theme:%x renderer:%d", env.WidthPx, env.ScaleMilli, env.TextScaleMilli, env.Locale, env.FontRevision, env.ThemeRevision, env.RendererRevision)
	}
	changed := env != p.env || len(history.Messages) != len(p.messages)
	if !changed && fresh {
		for i, m := range history.Messages {
			if m.Key != p.messages[i].Key || m.ContentRevision != p.messages[i].ContentRevision {
				changed = true
				break
			}
		}
	}
	if changed {
		if p.trace != nil {
			p.trace.History.Reason = historyInvalidationReason(p.env, env, len(p.messages), len(history.Messages))
		}
		// A message the composer sent, at the end of a history that shows its
		// end, flies to its place.
		var prevLast model.MessageID
		if n := len(p.messages); n > 0 && p.restored && !p.list.Position.BeforeEnd {
			prevLast = p.messages[n-1].Key.MessageID
		}
		p.rebuild(history.Messages, env)
		p.startFlights(gtx, prevLast, animate)
	}
	p.updateBot(c, history)
	v, anchored := p.source.Viewport(c.ID)
	anchored = anchored && !v.AtEnd
	// An anchor on its way, as when a search opens the chat at a message,
	// is waited for: restoring before it arrives would stay at another one.
	waiting := anchored && history.LoadingOlder && !p.hasMessage(v.AnchorMessageID)
	if !p.restored && len(p.messages) > 0 && !waiting {
		if anchored {
			p.restore(v.AnchorMessageID, v.AnchorOffsetPx)
			p.list.Position.BeforeEnd = true
		} else {
			p.list.Position.BeforeEnd = false
			if p.trace != nil {
				p.trace.History.Restore = "end"
				p.trace.Recorder.Event(p.trace.Window, "history.restore", "end", 0, 0)
			}
		}
		p.restored = true
	}
	p.finishJump(history)
	classic := p.classicComposer()
	// The floating composer covers the end of the history, which scrolls
	// out from under it; the classic one takes its height from the history.
	top, bottom, tail := 0, 0, gtx.Dp(80)
	// cover is how much of the history's bottom the composer hides.
	cover := 0
	if p.composer != nil && !classic {
		cover = floatingComposerCover(gtx, size)
	}
	if classic {
		bottom, tail = composerBarHeight(gtx, size), gtx.Dp(8)
	}
	if p.composer != nil && !p.frozen.Frozen() {
		// The strip of the message replied to takes room over the composer.
		reply := p.composer.replyHeight(gtx, c.ID, classic) + p.keyboardHeight(gtx, classic, size)
		if classic {
			bottom += reply
		} else {
			cover += reply
			tail += reply
		}
	}
	// The composer's middle is below the row's: a floating one is over the
	// tail of the history, a classic one under it.
	flightFrom := tail / 2
	if classic {
		flightFrom = tail + bottom/2
	}
	body := gtx
	body.Constraints = layout.Exact(image.Pt(size.X, max(0, size.Y-top-bottom)))
	if p.appearance != nil {
		// Dates and service messages lie on the wallpaper.
		body = p.appearance.historyContext(body)
	}
	offset(body, image.Pt(0, top), func(gtx layout.Context) layout.Dimensions {
		if len(p.messages) == 0 {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				gtx.Constraints.Min = image.Point{}
				if history.LoadingOlder {
					return p.loader.sized(gtx, l, 32)
				}
				return pill(gtx, l.T("history.empty"))
			})
		}
		p.selectionEvents(gtx)
		dims := p.list.Layout(gtx, len(p.messages), func(gtx layout.Context, i int) layout.Dimensions {
			msg := p.messages[i]
			date := p.dayStart[i]
			if p.trace != nil {
				p.trace.History.RowsLaidOut++
			}
			rowEnd := p.trace.Begin("history.rows")
			dims := p.flyingRow(gtx, msg.Key.MessageID, flightFrom, func(gtx layout.Context) layout.Dimensions {
				return p.row(gtx, msg, date, p.joins[i], l, animate)
			})
			if i == len(p.messages)-1 {
				dims.Size.Y += tail
			}
			rowEnd()
			if old, ok := p.measures[msg.Key.MessageID]; !ok || old.HeightPx != dims.Size.Y || old.ContentRevision != msg.ContentRevision {
				if p.trace != nil {
					p.trace.History.MeasurementsChanged++
				}
				measurement := model.MessageLayout{Key: msg.Key, Environment: env, ContentRevision: msg.ContentRevision, HeightPx: dims.Size.Y, MeasuredAt: time.Now()}
				p.measures[msg.Key.MessageID] = measurement
				p.dirty[msg.Key.MessageID] = measurement
				p.heights.Set(i, dims.Size.Y)
			}
			return dims
		})
		p.stickyAvatars(gtx, cover)
		p.stickyDate(gtx)
		p.selectionAreas(gtx)
		p.menuArea(gtx, top)
		return dims
	})
	var composerBackdrop *blurBackdrop
	if recording {
		call := record.Stop()
		call.Add(gtx.Ops)
		p.bd = newBackdrop(call, overlays.opacity)
		if p.blurComposer() {
			composerBackdrop = p.bd
		}
	}
	end := size.Y
	if p.composer != nil {
		p.composer.Layout(gtx, c.ID, l, p, animate, composerBackdrop)
		end = p.composer.top
	}
	p.jumpButtons(gtx, size, end, max(0, size.Y-top-bottom), history, l)
	p.errorMu.Lock()
	if err := p.mediaError; err != nil {
		p.mediaError = nil
		p.toast.Show(mediaErrorText(err))
	}
	p.errorMu.Unlock()
	p.botUpdate(l)
	p.toast.bd = p.toastBackdrop()
	p.toast.Layout(gtx, image.Rect(0, top, size.X, end))
	// Audio of a format no decoder here takes goes to the external player.
	if m := p.audio.takeExternal(); m != nil {
		p.play(gtx, *m, p.reportMedia, l)
	}
	p.menuLayout(gtx, l)
	if p.restored && p.list.Position.First < 3 && history.HasOlder && !history.LoadingOlder && history.Err == nil {
		p.source.LoadOlder(p.chat)
	}
	if p.restored && p.list.Position.First+p.list.Position.Count >= len(p.messages)-2 && history.HasNewer && !history.LoadingNewer && history.Err == nil {
		p.source.LoadNewer(p.chat)
	}
	p.save(false)
	if g, ok := p.source.(model.GhostStore); ok && (p.threadRoot == 0 || p.topic) {
		// What the history shows is read, if Ghost allows telling that.
		if id := p.bottomMessage(); id > 0 {
			g.MarkRead(c.ID, id, false)
		}
	}
	if len(p.dirty) > 0 {
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(500 * time.Millisecond)})
		if p.trace != nil {
			p.trace.Recorder.Count(p.trace.Window, "history.invalidate.pending-measurements")
		}
	}
	return layout.Dimensions{Size: size}
}

// layoutDialogs draws the page's dialogs over all of it, the header of
// the chat included: see layoutChatPage.
func (p *chatPage) layoutDialogs(gtx layout.Context, l localization.Catalog) {
	if p.link != "" {
		p.linkDialog(gtx, l)
	}
	p.tgLinkDialog(gtx, l)
	p.deleteDialog(gtx, l)
	p.membershipDialog(gtx, l)
	p.playerDialog(gtx, l)
	p.forwardDialog(gtx, l)
	p.emojiPacks.layout(gtx, p, l)
	p.stickers.layout(gtx, p, l)
	p.reacted.layout(gtx, p, l)
	p.edits.layout(gtx, p, l)
	p.shot.layout(gtx, p, l)
	p.translation.layout(gtx, p, l)
	if p.composer != nil {
		p.composer.layoutConfirm(gtx, p, l)
		p.composer.layoutFilesBox(gtx, p, l)
	}
}

// hasMessage reports whether the history shows the message.
func (p *chatPage) hasMessage(id model.MessageID) bool {
	for _, m := range p.messages {
		if m.Key.MessageID == id {
			return true
		}
	}
	return false
}

func (p *chatPage) restore(id model.MessageID, offset int) {
	var start time.Time
	scanned := 0
	if p.trace != nil {
		start = time.Now()
		defer func() {
			p.trace.History.Restore = "linear-id-anchor"
			p.trace.History.RestoreScanned += scanned
			p.trace.History.RestoreTime += time.Since(start)
			p.trace.Recorder.Event(p.trace.Window, "history.restore", fmt.Sprintf("chat:%d anchor:%d offset:%d linear scan", p.chat, id, offset), time.Since(start), scanned)
		}()
	}
	index := 0
	for i, m := range p.messages {
		scanned++
		if m.Key.MessageID <= id {
			index = i
		}
		if m.Key.MessageID == id {
			break
		}
	}
	p.list.Position.First = index
	p.list.Position.Offset = max(offset, 0)
}
func (p *chatPage) rebuild(messages []model.Message, env model.RenderEnvironment) {
	if p.trace != nil {
		start := time.Now()
		p.trace.History.Rebuilt = true
		defer func() {
			p.trace.History.RebuildTime = time.Since(start)
			p.trace.Recorder.Event(p.trace.Window, "history.rebuild", p.trace.History.Reason, time.Since(start), len(messages))
		}()
	}
	var anchor model.MessageID
	off := p.list.Position.Offset
	end := !p.list.Position.BeforeEnd
	if p.list.Position.First < len(p.messages) {
		anchor = p.messages[p.list.Position.First].Key.MessageID
	}
	if env != p.env || p.measures == nil {
		p.save(true)
		p.measures = map[model.MessageID]model.MessageLayout{}
		cacheEnd := p.trace.Begin("history.read-layout-cache")
		for _, l := range p.source.Layouts(p.chat, env) {
			p.measures[l.Key.MessageID] = l
		}
		cacheEnd()
		p.dirty = map[model.MessageID]model.MessageLayout{}
	}
	p.env = env
	p.messages = messages
	p.dates = make([]string, len(messages))
	p.dayStart = make([]bool, len(messages))
	p.joins = messageJoins(messages)
	p.nextDay = make([]int, len(messages))
	next := len(messages)
	for i := len(messages) - 1; i >= 0; i-- {
		p.dates[i] = messages[i].Date.Local().Format("02.01.2006")
		p.dayStart[i] = i == 0 || !sameDay(messages[i].Date, messages[i-1].Date)
		p.nextDay[i] = next
		if p.dayStart[i] {
			next = i
		}
	}

	hs := make([]int, len(messages))
	alive := map[model.MessageID]bool{}
	if p.textRunes == nil {
		p.textRunes = map[model.MessageID]textRunes{}
	}
	for i, m := range messages {
		alive[m.Key.MessageID] = true
		if l, ok := p.measures[m.Key.MessageID]; ok && l.ContentRevision == m.ContentRevision {
			hs[i] = l.HeightPx
			if p.trace != nil {
				p.trace.History.LayoutHits++
			}
			continue
		}
		runes, ok := p.textRunes[m.Key.MessageID]
		if !ok || runes.revision != m.ContentRevision {
			runes = textRunes{m.ContentRevision, utf8.RuneCountInString(m.Text)}
			p.textRunes[m.Key.MessageID] = runes
		}
		hs[i] = max(56, int(float32(60+runes.n/55*20)*float32(env.ScaleMilli)/1000))
		if m.Media != nil {
			hs[i] += 240
		}
	}
	for id := range p.rows {
		if !alive[id] {
			delete(p.rows, id)
		}
	}
	for id := range p.textRunes {
		if !alive[id] {
			delete(p.textRunes, id)
		}
	}
	p.pruneSelection(alive)
	p.heights = model.NewHeightIndex(hs)
	p.list.Measurements = p.heights
	if anchor != 0 && !end {
		p.restore(anchor, off)
	}
	p.list.Position.BeforeEnd = !end
}

// textRunes is how many runes a message's text has, at a revision.
type textRunes struct {
	revision uint64
	n        int
}

func (p *chatPage) richText(gtx layout.Context, r *messageRow, l localization.Catalog, animate bool) layout.Dimensions {
	end := p.trace.Begin("history.rich-text")
	defer end()
	p.textEvents(gtx, r, animate)
	theme := wdk.GetMaterialTheme(gtx)
	ty := theme.Typescale[token.TypestyleBodyLarge]
	styles := make([]styledtext.SpanStyle, len(r.runs))
	frames := make([]image.Image, len(r.runs))
	for i, run := range r.runs {
		st := styledtext.SpanStyle{Font: font.Font{Typeface: ty.Font, Weight: font.Normal}, Size: ty.Size, Content: run.Text, Color: scheme(gtx).Surface.OnColor.AsNRGBA()}
		if run.Bold {
			st.Font.Weight = font.Bold
		}
		if run.Italic {
			st.Font.Style = font.Italic
		}
		if run.Code {
			st.Font.Typeface = theme.Typescale[token.TypestylePreformatted].Font
		}
		if run.URL != "" {
			st.Color = scheme(gtx).Primary.Color.AsNRGBA()
		}
		if run.Quote {
			st.Font.Style = font.Italic
			st.Color = scheme(gtx).SurfaceVariant.OnColor.AsNRGBA()
		}
		if run.Emoji != 0 && (!run.Spoiler || r.revealed || !r.text.reveal.started.IsZero()) {
			msg := model.Message{Kind: model.MessageSticker, Media: &model.MessageMedia{ID: fmt.Sprintf("emoji/%d", run.Emoji), MIMEType: "application/x-custom-emoji"}}
			frames[i], _ = p.media.Frame(msg, animate)
			if frames[i] != nil {
				st.Color.A = 0
			}
		}
		styles[i] = st
	}
	r.text.fragments = r.text.fragments[:0]
	text := styledtext.Text(theme.TextShaper, styles...)
	text.Clusters = &r.text.clusters
	text.Decorate = func(gtx layout.Context, f styledtext.Fragment, draw func()) {
		r.text.fragments = append(r.text.fragments, f)
		i := f.Index
		run := r.runs[i]
		size := f.Bounds.Size()
		paintContent := func() {
			draw()
			if frames[i] != nil {
				drawImage(gtx, p.images, frames[i], size)
			}
			if run.Underline || run.Strike || run.URL != "" {
				y := size.Y - 1
				if run.Strike {
					y = size.Y / 2
				}
				paint.FillShape(gtx.Ops, styles[i].Color, clip.Rect(image.Rect(0, y, size.X, y+max(gtx.Dp(1), 1))).Op())
			}
		}
		if run.Spoiler && !r.revealed {
			radius := float32(0)
			if !r.text.reveal.started.IsZero() {
				radius = r.text.reveal.radius(gtx.Now, r.text.size)
			}
			center := r.text.reveal.center.Sub(f32.Pt(float32(f.Bounds.Min.X), float32(f.Bounds.Min.Y)))
			paintSpoiler(gtx, size, center, radius, scheme(gtx).OutlineVariant.AsNRGBA(), paintContent)
		} else {
			paintContent()
		}
	}
	record := op.Record(gtx.Ops)
	dims := text.Layout(gtx, nil)
	call := record.Stop()
	if p.activeText == r {
		for _, rect := range r.text.selectionRegions() {
			paint.FillShape(gtx.Ops, scheme(gtx).Primary.Color.SetOpacity(.28).AsNRGBA(), clip.Rect(rect).Op())
		}
	}
	call.Add(gtx.Ops)
	r.text.size = dims.Size
	area := clip.Rect{Max: dims.Size}.Push(gtx.Ops)
	pointer.CursorText.Add(gtx.Ops)
	r.text.clicker.Add(gtx.Ops)
	r.text.dragger.Add(gtx.Ops)
	area.Pop()
	return dims
}

func drawImage(gtx layout.Context, images *imageOps, im image.Image, size image.Point) layout.Dimensions {
	gtx.Constraints = layout.Exact(size)
	return widget.Image{Src: images.Op(im), Fit: widget.Contain}.Layout(gtx)
}
func safeURL(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, e := url.Parse(raw)
	if e != nil || u.Host == "" || u.User != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", false
	}
	return u.String(), true
}
func (p *chatPage) askLink(raw string) {
	if link, ok := model.ParseTelegramLink(raw); ok && p.openTelegramLink(link) {
		return
	}
	if target, ok := safeURL(raw); ok {
		p.link = target
		p.linkModal.Open()
	}
}
func (p *chatPage) linkDialog(gtx layout.Context, l localization.Catalog) {
	if p.cancel.Clicked(gtx) {
		p.linkModal.Close()
	}
	if p.open.Clicked(gtx) {
		target := p.link
		p.linkModal.Close()
		go func() {
			if err := openBrowser(target); err != nil {
				p.errorMu.Lock()
				p.mediaError = err
				p.errorMu.Unlock()
				p.invalidate()
			}
		}()
	}
	shown := p.linkModal.Layout(gtx, false, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(480))
		return card(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return label(gtx, l.T("history.external"), token.TypestyleTitleMedium, scheme(gtx).Surface.OnColor, 2)
			}), vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return label(gtx, p.link, token.TypestyleBodyMedium, scheme(gtx).Surface.OnColor, 8)
			}), vspace(12), layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx, layout.Rigid(func(gtx layout.Context) layout.Dimensions { return textButton(gtx, &p.cancel, l.T("history.cancel")) }), layout.Rigid(func(gtx layout.Context) layout.Dimensions { return textButton(gtx, &p.open, l.T("history.link")) }))
			}))
		}, 20)
	})
	if !shown {
		p.link = ""
	}
}

func historyInvalidationReason(old, next model.RenderEnvironment, previous, current int) string {
	var reasons []string
	if old.WidthPx != next.WidthPx {
		reasons = append(reasons, "width")
	}
	if old.ScaleMilli != next.ScaleMilli || old.TextScaleMilli != next.TextScaleMilli {
		reasons = append(reasons, "dpi/text-scale")
	}
	if old.Locale != next.Locale {
		reasons = append(reasons, "locale")
	}
	if old.FontRevision != next.FontRevision {
		reasons = append(reasons, "font")
	}
	if old.ThemeRevision != next.ThemeRevision {
		reasons = append(reasons, "theme")
	}
	if old.RendererRevision != next.RendererRevision {
		reasons = append(reasons, "renderer")
	}
	if previous != current {
		reasons = append(reasons, "message-count")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "message-key/content-revision")
	}
	return strings.Join(reasons, ",")
}
