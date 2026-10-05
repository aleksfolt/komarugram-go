// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	TelegramKeys["text.copy_code"] = "lng_code_block_header_copy"
	TelegramKeys["text.copied"] = "lng_text_copied"
	TelegramKeys["rich.photo"] = "lng_in_dlg_photo"
	TelegramKeys["rich.video"] = "lng_in_dlg_video"
	TelegramKeys["rich.album"] = "lng_in_dlg_album"
	TelegramKeys["rich.audio"] = "lng_in_dlg_audio_file"
	TelegramKeys["rich.file"] = "lng_in_dlg_file"
	TelegramKeys["rich.map"] = "lng_maps_point"
	TelegramKeys["rich.table"] = "lng_in_dlg_table"
	for key, texts := range map[string][2]string{
		"text.copy_code": {"Копировать", "Copy"},
		"text.copied":    {"Текст скопирован", "Text copied"},
		// Telegram uses icons for these actions; these are accessibility labels
		// and visible text for our shared text buttons.
		"text.expand_quote":   {"Развернуть цитату", "Expand quote"},
		"text.collapse_quote": {"Свернуть цитату", "Collapse quote"},
		// What a rich message without text shows (model.RichPage.Fallback).
		"rich.photo": {"Фото", "Photo"},
		"rich.video": {"Видео", "Video"},
		"rich.album": {"Альбом", "Album"},
		"rich.audio": {"Аудиофайл", "Audio file"},
		"rich.file":  {"Файл", "File"},
		"rich.map":   {"Геопозиция", "Location"},
		"rich.table": {"Таблица", "Table"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
