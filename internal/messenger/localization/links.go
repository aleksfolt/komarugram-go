// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	for key, text := range map[string][2]string{
		"username_not_found": {"Пользователь {user} не найден.", "User {user} not found."},
		"post_invalid":       {"К сожалению, эта публикация вам недоступна.", "Sorry, this post is unavailable."},
		"invite_bad":         {"К сожалению, ссылка-приглашение недействительна или устарела.", "Sorry, this invite link is invalid or has expired."},
		"join_group":         {"Вступить в группу", "Join Group"},
		"join_channel":       {"Подписаться", "Join Channel"},
		"request":            {"Подать заявку", "Request to Join"},
	} {
		russian["links."+key], english["links."+key] = text[0], text[1]
	}
	for key, tl := range map[string]string{"username_not_found": "lng_username_not_found", "post_invalid": "lng_error_post_link_invalid", "invite_bad": "lng_group_invite_bad_link", "join_group": "lng_profile_join_group", "join_channel": "lng_profile_join_channel", "request": "lng_group_request_to_join"} {
		TelegramKeys["links."+key] = tl
	}
}
