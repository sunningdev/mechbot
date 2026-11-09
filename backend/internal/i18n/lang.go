// Package i18n holds every message users can receive from the bot, in the same languages as the website.
package i18n

import "strings"

type Lang string

const (
	ZH   Lang = "zh"
	ZHTW Lang = "zh-TW"
	EN   Lang = "en"
	RU   Lang = "ru"
	ES   Lang = "es"
	FR   Lang = "fr"
	PT   Lang = "pt"
	VI   Lang = "vi"
)

// All lists the languages in the order the /lang picker shows them; the codes match the website's.
var All = []Lang{EN, ZH, ZHTW, RU, ES, FR, PT, VI}

// Parse accepts only the website's language codes.
func Parse(code string) (Lang, bool) {
	for _, l := range All {
		if string(l) == code {
			return l, true
		}
	}
	return "", false
}

// FromTelegram maps a Telegram client language (IETF tag such as "pt-br" or "zh-hant") to a supported language,
// falling back to English.
func FromTelegram(code string) Lang {
	code = strings.ToLower(strings.TrimSpace(code))
	switch {
	case code == "zh-hant" || code == "zh-tw" || code == "zh-hk" || code == "zh-mo":
		return ZHTW
	case strings.HasPrefix(code, "zh"):
		return ZH
	}
	base, _, _ := strings.Cut(code, "-")
	switch base {
	case "ru":
		return RU
	case "es":
		return ES
	case "fr":
		return FR
	case "pt":
		return PT
	case "vi":
		return VI
	}
	return EN
}

// Resolve picks the user's language: an explicit choice (website or /lang) first, then the Telegram client language.
// Users we know nothing about keep the original Chinese messages.
func Resolve(chosen, telegram string) Lang {
	if l, ok := Parse(chosen); ok {
		return l
	}
	if telegram != "" {
		return FromTelegram(telegram)
	}
	return ZH
}

// Name is how each language labels itself in the picker.
func (l Lang) Name() string {
	switch l {
	case ZH:
		return "简体中文"
	case ZHTW:
		return "繁體中文"
	case RU:
		return "Русский"
	case ES:
		return "Español"
	case FR:
		return "Français"
	case PT:
		return "Português"
	case VI:
		return "Tiếng Việt"
	}
	return "English"
}

// T returns the message set for l, falling back to English.
func T(l Lang) *Text {
	if t, ok := texts[l]; ok {
		return t
	}
	return texts[EN]
}

var texts = map[Lang]*Text{ZH: &zh, ZHTW: &zhTW, EN: &en, RU: &ru, ES: &es, FR: &fr, PT: &pt, VI: &vi}
