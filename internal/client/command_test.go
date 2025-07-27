package client

import "testing"

func TestGetCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "no args",
			args: []string{},
			want: "",
		},
		{
			name: "only program name",
			args: []string{"gophkeeper"},
			want: "",
		},
		{
			name: "command present",
			args: []string{"gophkeeper", CommandRegister},
			want: CommandRegister,
		},
		{
			name: "command present with extra args",
			args: []string{"gophkeeper", CommandLogin, "--username", "alice"},
			want: CommandLogin,
		},
		{
			name: "command present but empty string",
			args: []string{"gophkeeper", ""},
			want: "",
		},
		{
			name: "command is help",
			args: []string{"gophkeeper", CommandHelp},
			want: CommandHelp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetCommand(tt.args)
			if got != tt.want {
				t.Errorf("GetCommand(%v) = %q; want %q", tt.args, got, tt.want)
			}
		})
	}
}
