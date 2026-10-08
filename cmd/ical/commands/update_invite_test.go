package commands

import (
	"strings"
	"testing"
)

func TestUpdateInviteValidation(t *testing.T) {
	for _, tt := range []struct {
		name string
		args []string
		want string
	}{
		{"invalid address", []string{"--invite", "invalid"}, "invalid --invite value"},
		{"validate every address", []string{"--invite", "Eva <eva@example.com>", "--invite", "invalid"}, "invalid --invite value"},
		{"interactive", []string{"--invite", "eva@example.com", "-i"}, "--invite is not supported in interactive mode"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Cleanup(func() {
				updateInteractive = false
				updateInvite = nil
				for _, name := range []string{"invite", "interactive"} {
					flag := updateCmd.Flags().Lookup(name)
					if flag != nil {
						flag.Changed = false
					}
				}
			})
			if err := updateCmd.ParseFlags(tt.args); err != nil {
				t.Fatal(err)
			}
			err := updateCmd.RunE(updateCmd, []string{"unused-id"})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want %q", err, tt.want)
			}
		})
	}
}
