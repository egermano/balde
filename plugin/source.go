package plugin

import (
	"net/url"
	"strings"
)

// ResolveSource normalizes a plugin source reference. GitHub shorthand
// (github.com/owner/repo) becomes an https clone URL; everything else —
// full URLs, ssh remotes, local paths — passes through untouched.
func ResolveSource(source string) string {
	resolved, _ := ParseSourceReference(source)
	return resolved
}

// ParseSourceReference returns a clone URL/path and an optional revision.
// GitHub shorthand accepts github.com/owner/repo@tag; local paths also accept
// path@revision so callers can test and pin local repositories consistently.
func ParseSourceReference(source string) (string, string) {
	const host = "github.com"

	if !strings.HasPrefix(source, host+"/") {
		if at := strings.LastIndex(source, "@"); at > 0 &&
			!strings.Contains(source[:at], "://") && !strings.HasPrefix(source, "git@") {
			return source[:at], source[at+1:]
		}
		return source, ""
	}
	rest := strings.TrimPrefix(source, host+"/")
	ref := ""
	if at := strings.LastIndex(rest, "@"); at >= 0 {
		ref = rest[at+1:]
		rest = rest[:at]
		if ref == "" {
			return source, ""
		}
	}
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return source, ""
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return source, ""
		}
	}
	if strings.ContainsAny(rest, ":@") || strings.HasPrefix(rest, "/") {
		// not a plain owner/repo shape; leave it to git
		return source, ""
	}
	repo := strings.TrimSuffix(parts[1], ".git")
	return "https://" + host + "/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(repo) + ".git", ref
}
