package i18n

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

var verb = regexp.MustCompile(`%[-+# 0]*[0-9]*(\.[0-9]+)?[a-zA-Z%]`)

func TestEveryLanguageMatchesChinese(t *testing.T) {
	ref := reflect.ValueOf(zh)
	typ := ref.Type()
	for _, l := range All {
		txt, ok := texts[l]
		if !ok {
			t.Fatalf("%s: no texts", l)
		}
		v := reflect.ValueOf(*txt)
		for i := 0; i < typ.NumField(); i++ {
			name := typ.Field(i).Name
			got, want := v.Field(i).String(), ref.Field(i).String()
			if strings.TrimSpace(got) == "" {
				t.Errorf("%s.%s is empty", l, name)
				continue
			}
			if g, w := verb.FindAllString(got, -1), verb.FindAllString(want, -1); !reflect.DeepEqual(g, w) {
				t.Errorf("%s.%s verbs %v, zh has %v", l, name, g, w)
			}
		}
	}
}

func TestResolve(t *testing.T) {
	cases := []struct {
		chosen, tg string
		want       Lang
	}{
		{"", "", ZH},
		{"fr", "ru", FR},
		{"bogus", "pt-br", PT},
		{"", "zh-hant", ZHTW},
		{"", "zh-hans", ZH},
		{"", "de", EN},
		{"zh-TW", "", ZHTW},
	}
	for _, c := range cases {
		if got := Resolve(c.chosen, c.tg); got != c.want {
			t.Errorf("Resolve(%q, %q) = %s, want %s", c.chosen, c.tg, got, c.want)
		}
	}
}
