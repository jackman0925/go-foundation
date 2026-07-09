package netx

import (
	"fmt"
	"net/url"
	"strings"
)

// Domain 从 URL 字符串中返回 scheme://host[:port]。
func Domain(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Hostname() == "" {
		return "", fmt.Errorf("url must include scheme and host")
	}

	result := parsed.Scheme + "://" + parsed.Hostname()
	if port := parsed.Port(); port != "" {
		result += ":" + port
	}
	return result, nil
}

// URLPathJoin 拼接 URL path 片段，并保留第一个 scheme 和 host。
func URLPathJoin(parts ...string) (string, error) {
	if len(parts) == 0 {
		return "", nil
	}

	var scheme string
	var host string
	var query string
	var fragment string
	pathParts := make([]string, 0, len(parts))

	for _, part := range parts {
		parsed, err := url.Parse(part)
		if err != nil {
			return "", err
		}
		if scheme == "" {
			scheme = parsed.Scheme
		}
		if host == "" {
			host = parsed.Host
		}
		if parsed.RawQuery != "" {
			query = parsed.RawQuery
		}
		if parsed.Fragment != "" {
			fragment = parsed.Fragment
		}
		trimmed := strings.Trim(parsed.Path, "/")
		if trimmed != "" {
			pathParts = append(pathParts, trimmed)
		}
	}

	result := url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     "/" + strings.Join(pathParts, "/"),
		RawQuery: query,
		Fragment: fragment,
	}
	return result.String(), nil
}
