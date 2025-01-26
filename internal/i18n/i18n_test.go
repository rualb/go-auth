package i18n

import "testing"

func Test_userLang_Lang(t *testing.T) {
	type args struct {
		text string
		args []any
	}
	tests := []struct {
		name string
		x    *userLang
		args args
		want string
	}{
		// TODO: Add test cases.

		{
			name: "test with single argument",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "hello, {0}!",
				args: []any{"world"},
			},
			want: "hello, world!",
		},
		{
			name: "test with multiple arguments",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "{0} is {1} years old.",
				args: []any{"alice", 30},
			},
			want: "alice is 30 years old.",
		},
		{
			name: "test with no arguments",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "no arguments here.",
				args: []any{},
			},
			want: "no arguments here.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.Lang(tt.args.text, tt.args.args...); got != tt.want {
				t.Errorf("userLang.Lang() = %v, want %v", got, tt.want)
			}
		})
	}
}
