package main

import "testing"

func TestAutoStartValueMatches(t *testing.T) {
	exe := `C:\Program Files\TarkovPilot\TarkovPilot.exe`
	cases := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "current", value: `"C:\Program Files\TarkovPilot\TarkovPilot.exe" --hidden`, want: true},
		{name: "legacy", value: `"C:\Program Files\TarkovPilot\TarkovPilot.exe"`, want: true},
		{name: "other executable", value: `"C:\TarkovPilot.exe" --hidden`, want: false},
		{name: "missing argument", value: `C:\Program Files\TarkovPilot\TarkovPilot.exe`, want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := autoStartValueMatches(tc.value, exe); got != tc.want {
				t.Fatalf("autoStartValueMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestHasLaunchArg(t *testing.T) {
	args := []string{"updated", "--hidden"}
	if !hasLaunchArg(args, "updated") {
		t.Fatal("updated argument was not found")
	}
	if !hasLaunchArg(args, "--hidden") {
		t.Fatal("hidden argument was not found")
	}
	if hasLaunchArg(args, "--other") {
		t.Fatal("unexpected argument was found")
	}
}
