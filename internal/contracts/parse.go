package contracts

import (
	"regexp"
	"sort"
	"strings"
)

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

// ParseEnvFile parses a config.env (KEY=VALUE per line). Blank lines and
// lines starting with # are ignored; a line without = is skipped.
func ParseEnvFile(text string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(text, "\n") {
		t := strings.TrimSpace(line)
		if t == "" || strings.HasPrefix(t, "#") {
			continue
		}
		i := strings.Index(t, "=")
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(t[:i])
		if key == "" {
			continue
		}
		out[key] = strings.TrimSpace(t[i+1:])
	}
	return out
}

// MergeMaps overlays overlay onto base; overlay wins on shared keys.
func MergeMaps(base, overlay map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(overlay))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range overlay {
		out[k] = v
	}
	return out
}

var (
	// Leading \s* absorbs indentation: unlike the TS source these lines
	// aren't pre-trimmed, so the literal `^-?\s*` alone won't match indented list items.
	reSecretKey = regexp.MustCompile(`^\s*-?\s*secretKey:\s*(.+)$`)
	reKey       = regexp.MustCompile(`^\s*key:\s*(.+)$`)
	reRemoteRef = regexp.MustCompile(`^remoteRef:\s*$`)
)

// ParseSecretRefs parses an external-secrets overlay into secretKey ->
// remoteRef.key. Both the JSON-Patch shape and a plain spec.data[] list put
// secretKey: before its remoteRef:/key: pair in document order.
func ParseSecretRefs(text string) map[string]string {
	out := map[string]string{}
	var pending string
	sawRemoteRef := false
	for _, line := range strings.Split(text, "\n") {
		if m := reSecretKey.FindStringSubmatch(line); m != nil {
			pending = stripQuotes(m[1])
			sawRemoteRef = false
			continue
		}
		if reRemoteRef.MatchString(strings.TrimSpace(line)) {
			sawRemoteRef = true
			continue
		}
		if m := reKey.FindStringSubmatch(line); m != nil && pending != "" && sawRemoteRef {
			out[pending] = stripQuotes(m[1])
			pending = ""
			sawRemoteRef = false
		}
	}
	return out
}

// SortedKeys returns m's keys in ascending order (stable table rendering).
func SortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
