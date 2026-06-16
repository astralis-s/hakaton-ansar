package app

import "testing"

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
		ok   bool
	}{
		{"ru с +7 и разделителями", "+7 928 000-00-00", "+79280000000", true},
		{"ru с 8 в скобках", "8 (928) 000 00 00", "+79280000000", true},
		{"ru 11 цифр с 8", "89280000000", "+79280000000", true},
		{"ru 11 цифр с 7", "79280000000", "+79280000000", true},
		{"ru 10 цифр без кода", "9280000000", "+79280000000", true},
		{"межд. с плюсом 11 цифр", "+1 202 555 0143", "+12025550143", true},
		{"межд. с плюсом 12 цифр", "+44 20 7946 0958", "+442079460958", true},
		{"слишком коротко", "12345", "", false},
		{"пусто", "", "", false},
		{"только буквы", "не телефон", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := normalizePhone(tt.in)
			if ok != tt.ok {
				t.Fatalf("normalizePhone(%q) ok = %v, want %v", tt.in, ok, tt.ok)
			}
			if got != tt.want {
				t.Fatalf("normalizePhone(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNormalizeName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"  Иванов   Иван  ", "Иванов Иван"},
		{"Ибрагимов\tИслам\nВисханович", "Ибрагимов Ислам Висханович"},
	}
	for _, tt := range tests {
		if got := normalizeName(tt.in); got != tt.want {
			t.Errorf("normalizeName(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestValidName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"полное ФИО", "Ибрагимов Ислам Висханович", true},
		{"фамилия и имя", "Иванов Иван", true},
		{"короткое из двух слов", "Ан Ли", true},
		{"одно слово", "Иван", false},
		{"команда", "/start", false},
		{"пусто", "", false},
		{"слишком длинно", string(make([]rune, 0)) + longString(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// длину проверяем уже на нормализованной строке, как в проде
			if got := validName(normalizeName(tt.in)); got != tt.want {
				t.Fatalf("validName(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func longString() string {
	r := make([]rune, maxNameRunes+10)
	for i := range r {
		if i == maxNameRunes/2 {
			r[i] = ' '
		} else {
			r[i] = 'а'
		}
	}
	return string(r)
}
