// SPDX-License-Identifier: Unlicense OR MIT

package mockstore

import (
	"strings"
	"unicode/utf16"

	"komarugram/internal/messenger/model"
)

// TextBlocksExample is shared by the live demo and the render checks.
func TextBlocksExample() (string, []model.Entity) {
	code := "func greet() {\n    fmt.Println(\"Привет, мир! 👋\")\n}"
	quote := "Цитата сохраняет форматирование.\nСсылка: Telegram и скрытый спойлер.\nТретья строка цитаты.\nЧетвёртая строка видна после раскрытия."
	text := "Код и цитаты в сообщении 👋\n" + code + "\n" + quote + "\nТекст после блоков можно выделить вместе с ними."
	entity := func(kind, part string) model.Entity {
		i := strings.Index(text, part)
		return model.Entity{Kind: kind, Offset: len(utf16.Encode([]rune(text[:i]))), Length: len(utf16.Encode([]rune(part)))}
	}
	pre, q := entity("pre", code), entity("quote", quote)
	pre.Language, q.Collapsed = "go", true
	link := entity("url", "Telegram")
	link.URL = "https://telegram.org"
	return text, []model.Entity{pre, q, entity("bold", "форматирование"), link, entity("spoiler", "спойлер")}
}
