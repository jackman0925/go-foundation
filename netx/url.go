package netx

import (
	"net/url"
	"strings"
)

// Domain returns scheme://host[:port] from a URL string.
func Domain(value string) (string, error) {
	parsed, err := url.Parse(value)
	if err != nil {
		return "", err
	}

	result := parsed.Scheme + "://" + parsed.Hostname()
	if port := parsed.Port(); port != "" {
		result += ":" + port
	}
	return result, nil
}

// URLPathJoin joins URL path parts while preserving the first scheme and host.
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
