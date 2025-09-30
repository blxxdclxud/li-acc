package pdf

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLastIndex(t *testing.T) {
	tests := []struct {
		name      string
		slice     []rune
		value     rune
		wantValue int
	}{
		{
			name:      "nil slice",
			slice:     nil,
			value:     'a',
			wantValue: -1,
		},
		{
			name:      "empty slice",
			slice:     []rune{},
			value:     'a',
			wantValue: -1,
		},
		{
			name:      "single element - match",
			slice:     []rune{'a'},
			value:     'a',
			wantValue: 0,
		},
		{
			name:      "single element - no match",
			slice:     []rune{'b'},
			value:     'a',
			wantValue: -1,
		},
		{
			name:      "multiple elements - one match",
			slice:     []rune("abcdef"),
			value:     'c',
			wantValue: 2,
		},
		{
			name:      "multiple elements - multiple matches",
			slice:     []rune("abcaac"),
			value:     'a',
			wantValue: 4, // last 'a'
		},
		{
			name:      "match is last element",
			slice:     []rune("xyz"),
			value:     'z',
			wantValue: 2,
		},
		{
			name:      "no match in non-empty slice",
			slice:     []rune("xyz"),
			value:     'a',
			wantValue: -1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := lastIndex(test.slice, test.value)
			require.Equal(t, test.wantValue, res)
		})
	}
}

func TestPrettifyCredentialsString(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		splitWord string
		want      string
	}{
		{
			name:      "real receipt example",
			text:      "ЛС: 123456; ФИО обучающегося: Зубенко Михаил Петрович; Назначение: 10a доп питание сент; КБК: 82100000000000000131; ОКТМО: 98790098",
			splitWord: "Назначение",
			want:      "ЛС: 123456;   ФИО обучающегося: Зубенко Михаил Петрович;\nНазначение: 10a доп питание сент;   КБК: 82100000000000000131;   ОКТМО: 98790098",
		},
		{
			name:      "split word not found",
			text:      "ЛС: 123; ФИО: Иванов Иван; КБК: 111; ОКТМО: 222",
			splitWord: "Назначение",
			want:      "ЛС: 123;   ФИО: Иванов Иван;   КБК: 111;   ОКТМО: 222",
		},
		{
			name:      "empty string",
			text:      "",
			splitWord: "Назначение",
			want:      "",
		},
		{
			name:      "multiple occurrences of splitWord",
			text:      "A; B; Назначение: test; Назначение: second; C; D",
			splitWord: "Назначение",
			want:      "A;   B;\nНазначение: test;   Назначение: second;   C;   D",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := prettifyCredentialsString(tt.text, tt.splitWord)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		name   string
		amount string
		want   string
	}{
		{
			name:   "integer without decimal",
			amount: "2000",
			want:   "Сумма: 2000 руб. 00 коп.",
		},
		{
			name:   "integer with comma",
			amount: "3200,80",
			want:   "Сумма: 3200 руб. 80 коп.",
		},
		{
			name:   "integer with dot",
			amount: "450.50",
			want:   "Сумма: 450 руб. 50 коп.",
		},
		{
			name:   "zero amount",
			amount: "0",
			want:   "Сумма: 0 руб. 00 коп.",
		},
		{
			name:   "decimal less than 10",
			amount: "100.05",
			want:   "Сумма: 100 руб. 05 коп.",
		},
		{
			name:   "decimal with comma less than 10",
			amount: "78,07",
			want:   "Сумма: 78 руб. 07 коп.",
		},
		{
			name:   "integer with only one decimal number",
			amount: "1200.3",
			want:   "Сумма: 1200 руб. 30 коп.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatAmount(tt.amount)
			require.Equal(t, tt.want, got)
		})
	}
}
