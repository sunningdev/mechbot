package store

import (
	"testing"

	"mechbot/internal/i18n"
)

func TestUserLanguage(t *testing.T) {
	st := openTest(t)
	if got := st.UserLang(5, "ru"); got != i18n.RU {
		t.Fatalf("unknown user should follow Telegram: %s", got)
	}
	mustUser(t, st, 5, nil)
	if got := st.UserLang(5, ""); got != i18n.ZH {
		t.Fatalf("existing users without a language keep Chinese: %s", got)
	}
	if err := st.SetTelegramLang(5, "pt-br"); err != nil {
		t.Fatal(err)
	}
	if err := st.AddAddress(5, "0xaaa", 3); err != nil {
		t.Fatal(err)
	}
	watches, err := st.ListAddresses(5)
	if err != nil || len(watches) != 1 || watches[0].Lang != i18n.PT {
		t.Fatalf("watch should carry the Telegram language: %+v %v", watches, err)
	}
	if err := st.SetLang(5, i18n.FR); err != nil {
		t.Fatal(err)
	}
	u, err := st.GetUser(5)
	if err != nil || u.Lang != i18n.FR {
		t.Fatalf("a picked language wins over Telegram: %+v %v", u.Lang, err)
	}
	if got := st.UserLang(5, "ru"); got != i18n.FR {
		t.Fatalf("UserLang: %s", got)
	}
}
