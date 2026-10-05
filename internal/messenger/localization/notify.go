// SPDX-License-Identifier: Unlicense OR MIT

package localization

// Texts of notifications and their settings, as Telegram Desktop's
// Notifications and Sounds.
func init() {
	for key, texts := range map[string][2]string{
		"settings.notify":          {"Уведомления и звуки", "Notifications and Sounds"},
		"notify.global":            {"Общие настройки", "Global settings"},
		"notify.desktop":           {"Уведомления на рабочем столе", "Desktop notifications"},
		"notify.sound":             {"Разрешить звук", "Allow sound"},
		"notify.shown":             {"Показывать в уведомлении", "Show in notifications"},
		"notify.name":              {"Имя", "Name"},
		"notify.text":              {"Текст", "Text"},
		"notify.chats":             {"Уведомления из чатов", "Notifications for chats"},
		"notify.private":           {"Личные чаты", "Private chats"},
		"notify.groups":            {"Группы", "Groups"},
		"notify.channels":          {"Каналы", "Channels"},
		"notify.chats_hint":        {"Чаты, где уведомления выключены, молчат всегда.", "Muted chats stay silent."},
		"notify.from":              {"Показывать уведомления от", "Show notifications from"},
		"notify.all_accounts":      {"Всех аккаунтов", "All accounts"},
		"notify.all_accounts_hint": {"Отключите, если хотите получать уведомления только от аккаунта, которым пользуетесь сейчас.", "Turn this off if you want to receive notifications only from the account you are currently using."},
		"notify.new_message":       {"У вас новое сообщение", "You have a new message"},
		"notify.on":                {"Включены", "On"},
		"notify.off":               {"Выключены", "Off"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
	for key, lng := range map[string]string{
		"settings.notify":          "lng_settings_section_notify",
		"notify.global":            "lng_settings_notify_global",
		"notify.desktop":           "lng_settings_desktop_notify",
		"notify.sound":             "lng_settings_sound_allowed",
		"notify.name":              "lng_notification_show_name",
		"notify.text":              "lng_notification_show_text",
		"notify.chats":             "lng_settings_notify_title",
		"notify.private":           "lng_notification_private_chats",
		"notify.groups":            "lng_notification_groups",
		"notify.channels":          "lng_notification_channels",
		"notify.from":              "lng_settings_show_from",
		"notify.all_accounts":      "lng_settings_notify_all",
		"notify.all_accounts_hint": "lng_settings_notify_all_about",
		"notify.new_message":       "lng_notification_preview",
	} {
		TelegramKeys[key] = lng
	}
}
