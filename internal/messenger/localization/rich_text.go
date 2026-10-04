// SPDX-License-Identifier: Unlicense OR MIT

package localization

func init() {
	TelegramKeys["text.copy_code"] = "lng_code_block_header_copy"
	TelegramKeys["text.copied"] = "lng_text_copied"
	for key, texts := range map[string][2]string{
		"text.copy_code": {"Копировать", "Copy"},
		"text.copied":    {"Текст скопирован", "Text copied"},
		// Telegram uses icons for these actions; these are accessibility labels
		// and visible text for our shared text buttons.
		"text.expand_quote":   {"Развернуть цитату", "Expand quote"},
		"text.collapse_quote": {"Свернуть цитату", "Collapse quote"},
	} {
		russian[key], english[key] = texts[0], texts[1]
	}
}
