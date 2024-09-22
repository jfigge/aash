/*
 * Copyright (C) 2024 by Jason Figge
 */

package utils

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"golang.org/x/term"
	"us.figge.auto-ssh/internal/core/config"
)

var (
	width int
)

func termWidth() int {
	if width == 0 {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return 80
		}
		defer func() {
			_ = term.Restore(int(os.Stdin.Fd()), oldState)
		}()
		if w, _, e := term.GetSize(int(os.Stdin.Fd())); e != nil {
			width = 80
		} else if w < 60 {
			width = 60
		} else {
			width = w
		}
	}
	return width
}

func TruncateText(s string, max int) string {
	if len(s) < max {
		return s
	}
	return s[:strings.LastIndexAny(s[:max-3], "\t\n .,:;-")] + "..."
}

func TruncateLine(s string, offset int) string {
	max := termWidth() - offset
	if len(s) < max {
		return s
	}
	return s[:strings.LastIndexAny(s[:max-3], "\t\n .,:;-")] + "..."
}

func Wrap(s string, indent int) string {
	max := termWidth() - indent
	var lines []string
	var padding = strings.Repeat(" ", indent)
	for len(s) > max {
		index := strings.LastIndexAny(s[:max], "\t\n .|,:;-")
		if index < (max/10) || index < 10 || index > max {
			index = len(s[:max])
		}
		line := s[:index]
		if len(lines) > 0 {
			line = padding + line
		}
		lines = append(lines, line)
		s = strings.TrimSpace(s[index:])
	}
	extra := "\n"
	if len(s) > 0 {
		if len(lines) > 0 {
			s = padding + s
		}
		lines = append(lines, s)
		extra = ""
	}
	if lines == nil {
		extra = ""
	}
	return strings.Join(lines, "\n") + extra
}

func VPrintf(template string, v ...interface{}) {
	if config.VerboseFlag {
		fmt.Printf(template, v...)
	}
}

func Askf(prompt string, hidden bool, inline bool, args ...any) (string, bool) {
	return Ask(fmt.Sprintf(prompt, args...), hidden, inline)
}
func Ask(prompt string, hidden bool, inline bool) (string, bool) {
	if config.ForcedFlag {
		return "yes", true
	}
	fmt.Print(prompt)
	if !inline {
		fmt.Println()
	}
	var text string
	if hidden {
		byteText, err := term.ReadPassword(int(os.Stdin.Fd()))
		if err != nil {
			return "", false
		}
		text = string(byteText)
	} else {
		_, err := fmt.Scanln(&text)
		if err != nil {
			return "", false
		}
	}
	text = strings.TrimSpace(text)
	return text, len(text) > 0
}

func Obfuscate(value interface{}) string {
	txt, ok := value.(string)
	if !ok || len(txt) < 10 {
		value = "***"
	} else if len(txt) < 20 {
		value = "***" + txt[len(txt)-3:]
	} else {
		value = txt[:2] + "***" + txt[len(txt)-3:]
	}
	return fmt.Sprintf("%v", value)
}

func ExpandHome(path string) string {
	path, err := ExpandHomeE(path)
	if config.VerboseFlag {
		fmt.Printf("failed to expand ~: %v\n", err)
	}
	return path
}

func ExpandHomeE(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return path, err
	}

	dir := usr.HomeDir
	return filepath.Join(dir, path[2:]), nil
}
