package config

import (
	"bufio"
	"io"
	"strings"
)

func ParseYAML(r io.Reader) map[string]string {
	out := map[string]string{}
	s := bufio.NewScanner(r)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.SplitN(line, ":", 2)
		if len(p) != 2 {
			continue
		}
		out[strings.TrimSpace(p[0])] = strings.Trim(strings.TrimSpace(p[1]), "\"")
	}
	return out
}
