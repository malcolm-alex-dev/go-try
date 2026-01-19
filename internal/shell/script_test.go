package shell

import (
	"strings"
	"testing"
)

func TestQuote(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"with'quote", "'with'\"'\"'quote'"},
		{"/path/to/dir", "'/path/to/dir'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := quote(tt.input)
			if got != tt.want {
				t.Errorf("quote(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestScriptCD(t *testing.T) {
	script := CD("/path/to/dir")

	if !strings.Contains(script, "touch '/path/to/dir'") {
		t.Error("script should contain touch command")
	}
	if !strings.Contains(script, "cd '/path/to/dir'") {
		t.Error("script should contain cd command")
	}
	if !strings.Contains(script, "# if you can read this") {
		t.Error("script should contain warning comment")
	}
}

func TestScriptMkdirCD(t *testing.T) {
	script := MkdirCD("/path/to/new")

	if !strings.Contains(script, "mkdir -p '/path/to/new'") {
		t.Error("script should contain mkdir command")
	}
	if !strings.Contains(script, "cd '/path/to/new'") {
		t.Error("script should contain cd command")
	}
}

func TestScriptClone(t *testing.T) {
	script := Clone("/path/to/dir", "git@github.com:user/repo.git")

	if !strings.Contains(script, "git clone") {
		t.Error("script should contain git clone command")
	}
	if !strings.Contains(script, "git@github.com:user/repo.git") {
		t.Error("script should contain git URL")
	}
	if !strings.Contains(script, "cd '/path/to/dir'") {
		t.Error("script should contain cd command")
	}
}

func TestScriptDelete(t *testing.T) {
	paths := []string{"/base/dir1", "/base/dir2"}
	script := Delete(paths, "/base")

	if !strings.Contains(script, "cd '/base'") {
		t.Error("script should cd to base first")
	}
	if !strings.Contains(script, "rm -rf") {
		t.Error("script should contain rm command")
	}
}

func TestInitBash(t *testing.T) {
	script := InitBash("/usr/local/bin/try", "/home/user/tries")

	if !strings.Contains(script, "try()") {
		t.Error("should define try function")
	}
	if !strings.Contains(script, "eval") {
		t.Error("should eval the output")
	}
	if !strings.Contains(script, "--path") {
		t.Error("should include path flag")
	}
}

func TestInitFish(t *testing.T) {
	script := InitFish("/usr/local/bin/try", "")

	if !strings.Contains(script, "function try") {
		t.Error("should define try function")
	}
	if !strings.Contains(script, "eval") {
		t.Error("should eval the output")
	}
}

func TestScriptBuilder(t *testing.T) {
	s := New().
		AddMkdir("/path").
		AddCD("/path")

	result := s.String()

	// Check commands are chained with &&
	if !strings.Contains(result, "&& \\") {
		t.Error("commands should be chained with && \\")
	}
}
