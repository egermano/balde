package plugin

import (
	"net/url"
	"strings"
)

// ResolveSource normalizes a plugin source reference. GitHub shorthand
// (github.com/owner/repo) becomes an https clone URL; everything else —
// full URLs, ssh remotes, local paths — passes through untouched.
func ResolveSource(source string) string {
	const host = "github.com"

	if !strings.HasPrefix(source, host+"/") {
		return source
	}
	rest := strings.TrimPrefix(source, host+"/")
	parts := strings.Split(rest, "/")
	if len(parts) != 2 {
		return source
	}
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return source
		}
	}
	if strings.ContainsAny(rest, ":@") || strings.HasPrefix(rest, "/") {
		// not a plain owner/repo shape; leave it to git
		return source
	}
	return "https://" + host + "/" + url.PathEscape(parts[0]) + "/" + url.PathEscape(parts[1]) + ".git"
}
