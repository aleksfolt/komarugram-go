package localization

import (
	"komarugram/internal/messenger/model"
	"strconv"
	"strings"
)

func init() {
	mappings := map[string]string{
		"chat_theme.import": "lng_theme_editor_menu_import", "viewer.position": "lng_mediaview_n_of_amount", "viewer.photo": "lng_mediaview_single_photo", "viewer.video": "lng_in_dlg_video", "page.choose_chat": "lng_bot_choose_chat",
		"history.select_forward": "lng_selected_forward", "history.select_delete": "lng_selected_delete", "history.reply": "lng_context_reply_msg", "history.forward": "lng_settings_forwards_privacy", "history.link": "lng_open_link", "shared.more": "lng_stories_show_more",
		"history.selection_count": "lng_media_selected_message",
		"status.bot":              "lng_status_bot", "status.recently": "lng_status_recently", "status.members": "lng_chat_status_members", "status.online": "lng_chat_status_online", "status.members_online": "lng_chat_status_members_online", "status.subscribers": "lng_chat_status_subscribers",
		"shared.photos": "lng_media_type_photos", "shared.videos": "lng_media_type_videos", "shared.files": "lng_media_type_files", "shared.music": "lng_media_type_songs", "shared.voice": "lng_media_type_audios", "shared.links": "lng_media_type_links", "shared.gifs": "lng_media_type_gifs", "shared.polls": "lng_media_type_polls", "shared.saved": "lng_saved_messages", "shared.stories": "lng_media_type_posts", "shared.gifts": "lng_media_type_gifts", "shared.groups": "lng_profile_common_groups_section",
		"history.loading": "lng_profile_loading", "history.retry": "lng_bot_download_retry", "history.cancel": "lng_cancel", "viewer.close": "lng_close", "settings.back": "lng_go_back", "file.open": "lng_markdown_preview_open_file", "poll.closed": "lng_polls_closed", "poll.open": "lng_polls_public", "poll.quiz": "lng_polls_public_quiz",
		"chat_theme.title": "lng_chat_theme_change", "chat_theme.default": "lng_filters_tabs_default", "chat_theme.apply": "lng_chat_theme_apply", "history.file": "lng_in_dlg_file", "history.gif": "lng_media_type_gifs",
		"gift.from": "lng_gift_link_label_from", "gift.date": "lng_gift_link_label_date", "gift.value": "lng_gift_link_label_value", "gift.model": "lng_gift_unique_model", "gift.symbol": "lng_gift_unique_symbol", "gift.backdrop": "lng_gift_unique_backdrop", "gift.quantity": "lng_gift_unique_availability_label", "gift.hidden_sender": "lng_gift_from_hidden", "gift.title": "lng_gift_link_label_gift", "gift.number": "lng_gift_unique_number", "gift.terms": "lng_credits_box_out_about", "gift.terms_link": "lng_credits_summary_options_about_link",
		"gift.stars": "lng_gift_stars_title", "gift.issued": "lng_gift_unique_availability", "poll.count": "lng_polls_votes_count",
	}
	for k, v := range mappings {
		TelegramKeys[k] = v
	}
	counters := map[model.SharedKind]string{model.SharedPhotos: "lng_profile_photos", model.SharedVideos: "lng_profile_videos", model.SharedFiles: "lng_profile_files", model.SharedMusic: "lng_profile_songs", model.SharedVoice: "lng_profile_audios", model.SharedLinks: "lng_profile_shared_links", model.SharedGIFs: "lng_profile_gifs", model.SharedPolls: "lng_profile_polls", model.SharedSaved: "lng_profile_saved_messages", model.SharedStories: "lng_profile_saved_stories", model.SharedGifts: "lng_profile_peer_gifts", model.SharedGroups: "lng_profile_common_groups"}
	for kind, key := range counters {
		app := "shared.count." + string(kind)
		TelegramKeys[app] = key
		russian[app] = "{count} · " + For("ru").T("shared."+string(kind))
		english[app] = "{count} · " + For("en").T("shared."+string(kind))
	}
	empty := map[string]string{"photos": "photo", "videos": "video", "files": "file", "music": "song", "voice": "audio", "links": "link", "gifs": "gif"}
	for k, v := range empty {
		TelegramKeys["shared.empty."+k] = "lng_media_" + v + "_empty"
		russian["shared.empty."+k] = russian["shared.empty"]
		english["shared.empty."+k] = english["shared.empty"]
	}
	ru := map[string]string{"viewer.position": "{n} из {amount}", "viewer.photo": "Фотография", "viewer.video": "Видео", "status.members": "{count} участников", "status.online": "{count} в сети", "status.members_online": "{members_count}, {online_count}", "status.subscribers": "{count} подписчиков", "history.selection_count": "Выбрано: {count}", "gift.stars": "{count} Stars", "gift.from": "Отправитель", "gift.date": "Дата", "gift.value": "Стоимость", "gift.model": "Модель", "gift.symbol": "Символ", "gift.backdrop": "Фон", "gift.quantity": "Количество", "gift.hidden_sender": "Скрытый отправитель", "gift.title": "Подарок", "gift.number": "Коллекционный №{index}", "gift.issued": "Выпущено {count} из {amount}", "gift.terms": "Ознакомьтесь с {link} Telegram Stars.", "gift.terms_link": "условиями использования", "poll.count": "Голосов: {count}"}
	en := map[string]string{"viewer.position": "{n} of {amount}", "viewer.photo": "Photo", "viewer.video": "Video", "status.members": "{count} members", "status.online": "{count} online", "status.members_online": "{members_count}, {online_count}", "status.subscribers": "{count} subscribers", "history.selection_count": "Selected: {count}", "gift.stars": "{count} Stars", "gift.from": "From", "gift.date": "Date", "gift.value": "Value", "gift.model": "Model", "gift.symbol": "Symbol", "gift.backdrop": "Backdrop", "gift.quantity": "Quantity", "gift.hidden_sender": "Hidden sender", "gift.title": "Gift", "gift.number": "Collectible #{index}", "gift.issued": "{count} of {amount} issued", "gift.terms": "Please review the Telegram Stars {link}.", "gift.terms_link": "Terms and Conditions", "poll.count": "Votes: {count}"}
	for k, v := range ru {
		russian[k] = v
	}
	for k, v := range en {
		english[k] = v
	}
	for key, forms := range map[string][3]string{"status.members": {"{count} участник", "{count} участника", "{count} участников"}, "status.subscribers": {"{count} подписчик", "{count} подписчика", "{count} подписчиков"}, "gift.stars": {"{count} звезда", "{count} звезды", "{count} звёзд"}, "poll.count": {"{count} голос", "{count} голоса", "{count} голосов"}} {
		russian[key+"#one"] = forms[0]
		russian[key+"#few"] = forms[1]
		russian[key+"#many"] = forms[2]
	}
	english["status.members#one"] = "{count} member"
	english["status.subscribers#one"] = "{count} subscriber"
	english["gift.stars#one"] = "{count} Star"
	english["poll.count#one"] = "{count} vote"

}

// Format substitutes named Telegram placeholders as data, never as printf syntax.
func (c Catalog) Format(key string, args map[string]string) string {
	return replaceTags(c.T(key), args)
}
func replaceTags(s string, args map[string]string) string {
	var pairs []string
	for k, v := range args {
		pairs = append(pairs, "{"+k+"}", v)
	}
	if len(pairs) > 0 {
		s = strings.NewReplacer(pairs...).Replace(s)
	}
	return s
}
func (c Catalog) Count(key string, n int, args map[string]string) string {
	category := "other"
	if c.language == Russian {
		switch {
		case n%10 == 1 && n%100 != 11:
			category = "one"
		case n%10 >= 2 && n%10 <= 4 && (n%100 < 12 || n%100 > 14):
			category = "few"
		default:
			category = "many"
		}
	} else if n == 1 {
		category = "one"
	}
	serverKey := TelegramKeys[key]
	value := c.server[serverKey+"#"+category]
	if value == "" {
		value = c.server[serverKey+"#other"]
	}
	if value == "" {
		local := russian
		if c.language == English {
			local = english
		}
		value = local[key+"#"+category]
		if value == "" {
			value = c.T(key)
		}
	}
	tags := map[string]string{"count": strconv.Itoa(n)}
	for k, v := range args {
		tags[k] = v
	}
	return replaceTags(value, tags)
}
func (c Catalog) SharedCount(kind model.SharedKind, n int) string {
	return c.Count("shared.count."+string(kind), n, nil)
}
func (c Catalog) SharedEmpty(kind model.SharedKind) string {
	key := "shared.empty." + string(kind)
	if _, ok := TelegramKeys[key]; ok {
		return c.T(key)
	}
	return c.T("shared.empty")
}
