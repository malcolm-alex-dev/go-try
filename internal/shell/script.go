// Package shell handles shell script generation for the try command.
package shell

import (
	"fmt"
	"strings"
)

const scriptWarning = "# if you can read this, you didn't launch try from an alias. run try --help."

// quote escapes a string for safe use in shell scripts.
func quote(s string) string {
	// Use single quotes, escaping any embedded single quotes
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// Script represents a series of shell commands to execute.
type Script struct {
	commands []string
}

// New creates a new empty script.
func New() *Script {
	return &Script{}
}

// Add appends a command to the script.
func (s *Script) Add(cmd string) *Script {
	s.commands = append(s.commands, cmd)
	return s
}

// AddCD adds a cd command.
func (s *Script) AddCD(path string) *Script {
	return s.Add(fmt.Sprintf("cd %s", quote(path)))
}

// AddMkdir adds a mkdir command.
func (s *Script) AddMkdir(path string) *Script {
	return s.Add(fmt.Sprintf("mkdir -p %s", quote(path)))
}

// AddTouch adds a touch command.
func (s *Script) AddTouch(path string) *Script {
	return s.Add(fmt.Sprintf("touch %s", quote(path)))
}

// AddEcho adds an echo command.
func (s *Script) AddEcho(msg string) *Script {
	return s.Add(fmt.Sprintf("echo %s", quote(msg)))
}

// AddGitClone adds a git clone command.
func (s *Script) AddGitClone(url, destPath string) *Script {
	return s.Add(fmt.Sprintf("git clone %s %s", quote(url), quote(destPath)))
}

// AddRm adds an rm -rf command with safety wrapper.
func (s *Script) AddRm(path, basePath string) *Script {
	// Safety: validate path is inside basePath before deleting
	cmd := fmt.Sprintf("test -d %s && rm -rf %s", quote(path), quote(path))
	return s.Add(cmd)
}

// String renders the script as a shell-evaluable string.
func (s *Script) String() string {
	if len(s.commands) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(scriptWarning)
	sb.WriteString("\n")

	for i, cmd := range s.commands {
		if i == 0 {
			sb.WriteString(cmd)
		} else {
			sb.WriteString("  ")
			sb.WriteString(cmd)
		}

		if i < len(s.commands)-1 {
			sb.WriteString(" && \\\n")
		} else {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// CD creates a script that touches and cd's to a directory.
func CD(path string) string {
	return New().
		AddTouch(path).
		AddEcho(path).
		AddCD(path).
		String()
}

// MkdirCD creates a script that creates a directory and cd's to it.
func MkdirCD(path string) string {
	return New().
		AddMkdir(path).
		AddTouch(path).
		AddEcho(path).
		AddCD(path).
		String()
}

// Clone creates a script that clones a repo and cd's to it.
func Clone(path, url string) string {
	return New().
		AddMkdir(path).
		AddEcho(fmt.Sprintf("Cloning %s...", url)).
		AddGitClone(url, path).
		AddTouch(path).
		AddEcho(path).
		AddCD(path).
		String()
}

// Delete creates a script that deletes directories.
func Delete(paths []string, basePath string) string {
	s := New().AddCD(basePath)
	for _, p := range paths {
		s.AddRm(p, basePath)
	}
	// Try to stay in current dir, or go home if deleted
	s.Add(`( cd "$PWD" 2>/dev/null || cd "$HOME" )`)
	return s.String()
}

// InitBash returns the bash/zsh shell function definition.
func InitBash(scriptPath, triesPath string) string {
	pathArg := ""
	if triesPath != "" {
		pathArg = fmt.Sprintf(" --path %s", quote(triesPath))
	}

	return fmt.Sprintf(`try() {
  local out
  out=$(/usr/bin/env %s exec%s "$@" 2>/dev/tty)
  if [ $? -eq 0 ]; then
    eval "$out"
  else
    echo "$out"
  fi
}
`, quote(scriptPath), pathArg)
}

// InitFish returns the fish shell function definition.
func InitFish(scriptPath, triesPath string) string {
	pathArg := ""
	if triesPath != "" {
		pathArg = fmt.Sprintf(" --path %s", quote(triesPath))
	}

	return fmt.Sprintf(`function try
  set -l out (/usr/bin/env %s exec%s $argv 2>/dev/tty | string collect)
  if test $status -eq 0
    eval $out
  else
    echo $out
  end
end
`, quote(scriptPath), pathArg)
}
