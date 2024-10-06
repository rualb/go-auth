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
			name: "Test with single argument",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "Hello, {0}!",
				args: []any{"World"},
			},
			want: "Hello, World!",
		},
		{
			name: "Test with multiple arguments",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "{0} is {1} years old.",
				args: []any{"Alice", 30},
			},
			want: "Alice is 30 years old.",
		},
		{
			name: "Test with no arguments",
			x:    &userLang{}, // Initialize with actual instance if needed
			args: args{
				text: "No arguments here.",
				args: []any{},
			},
			want: "No arguments here.",
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
