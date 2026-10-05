// SPDX-License-Identifier: Unlicense OR MIT

package model

import "testing"

func TestParseTelegramLink(t *testing.T) {
	for raw, want := range map[string]TelegramLink{
		"https://t.me/durov":                   {Kind: LinkUsername, Name: "durov"},
		"t.me/durov/42":                        {Kind: LinkUsername, Name: "durov", Post: 42},
		"https://telegram.me/gobot?start=ref1": {Kind: LinkUsername, Name: "gobot", Start: "ref1"},
		"https://t.me/forum/7/120":             {Kind: LinkUsername, Name: "forum", Post: 120},
		"https://durov.t.me":                   {Kind: LinkUsername, Name: "durov"},
		"https://t.me/s/telegram":              {Kind: LinkUsername, Name: "telegram"},
		"https://t.me/c/1234567/89":            {Kind: LinkPrivatePost, Channel: 1234567, Post: 89},
		"https://t.me/c/1234567/5/89":          {Kind: LinkPrivatePost, Channel: 1234567, Post: 89},
		"https://t.me/+AbCdEf_123":             {Kind: LinkInvite, Hash: "AbCdEf_123"},
		"https://t.me/joinchat/AbCdEf":         {Kind: LinkInvite, Hash: "AbCdEf"},
		"https://t.me/addstickers/Animals":     {Kind: LinkStickers, Name: "Animals"},
		"https://t.me/addemoji/Hearts":         {Kind: LinkEmoji, Name: "Hearts"},
		"tg://resolve?domain=durov&post=5":     {Kind: LinkUsername, Name: "durov", Post: 5},
		"tg://privatepost?channel=77&post=3":   {Kind: LinkPrivatePost, Channel: 77, Post: 3},
		"tg://addstickers?set=Animals":         {Kind: LinkStickers, Name: "Animals"},
		"tg://join?invite=AbCd":                {Kind: LinkInvite, Hash: "AbCd"},
		"tg://user?id=42":                      {Kind: LinkUser, User: 42},
	} {
		got, ok := ParseTelegramLink(raw)
		if !ok || got != want {
			t.Errorf("%s: %+v, %v; want %+v", raw, got, ok, want)
		}
	}
	for _, raw := range []string{
		"https://telegram.org", "https://t.me/", "https://t.me/share/url?url=x", "https://t.me/proxy?server=a",
		"https://t.me/ab", "https://t.me/durov/abc", "https://t.me/c/x/1", "https://example.com/t.me/durov",
		"tg://settings", "ftp://t.me/durov", "https://t.me/addstickers/", "https://t.me/a/b/c/d",
	} {
		if got, ok := ParseTelegramLink(raw); ok {
			t.Errorf("%s: %+v", raw, got)
		}
	}
}
