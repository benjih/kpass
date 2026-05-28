package args

import "testing"

func TestParseInitialArg(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantFilePath string
		wantFillHost string
	}{
		{
			name:         "no args",
			args:         nil,
			wantFilePath: "",
			wantFillHost: "",
		},
		{
			name:         "plain file path",
			args:         []string{"/home/user/db.kdbx"},
			wantFilePath: "/home/user/db.kdbx",
			wantFillHost: "",
		},
		{
			name:         "kpass fill URI",
			args:         []string{"kpass://fill?host=github.com"},
			wantFilePath: "",
			wantFillHost: "github.com",
		},
		{
			name:         "kpass fill URI with empty host param",
			args:         []string{"kpass://fill?host="},
			wantFilePath: "",
			wantFillHost: "",
		},
		{
			name:         "kpass URI with wrong path host",
			args:         []string{"kpass://other?host=github.com"},
			wantFilePath: "",
			wantFillHost: "",
		},
		{
			name:         "flags are skipped",
			args:         []string{"-v", "--debug", "/path/to/db.kdbx"},
			wantFilePath: "/path/to/db.kdbx",
			wantFillHost: "",
		},
		{
			name:         "flag before kpass URI",
			args:         []string{"-v", "kpass://fill?host=example.com"},
			wantFilePath: "",
			wantFillHost: "example.com",
		},
		{
			name:         "first non-flag file wins",
			args:         []string{"/first.kdbx", "/second.kdbx"},
			wantFilePath: "/first.kdbx",
			wantFillHost: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotHost := ParseInitialArg(tt.args)
			if gotFile != tt.wantFilePath || gotHost != tt.wantFillHost {
				t.Errorf("ParseInitialArg(%v) = (%q, %q), want (%q, %q)",
					tt.args, gotFile, gotHost, tt.wantFilePath, tt.wantFillHost)
			}
		})
	}
}
