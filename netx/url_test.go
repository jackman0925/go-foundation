package netx

import "testing"

func TestDomain(t *testing.T) {
	got, err := Domain("https://example.com:8443/a/b?x=1")
	if err != nil {
		t.Fatalf("Domain returned error: %v", err)
	}
	if got != "https://example.com:8443" {
		t.Fatalf("unexpected domain: %q", got)
	}
}

func TestDomainRejectsMissingSchemeOrHost(t *testing.T) {
	if _, err := Domain("example.com/path"); err == nil {
		t.Fatal("expected missing scheme or host error")
	}
	if _, err := Domain("https:///path"); err == nil {
		t.Fatal("expected missing host error")
	}
}

func TestURLPathJoin(t *testing.T) {
	got, err := URLPathJoin("https://example.com/api/", "/v1/", "users?active=true")
	if err != nil {
		t.Fatalf("URLPathJoin returned error: %v", err)
	}
	if got != "https://example.com/api/v1/users?active=true" {
		t.Fatalf("unexpected URL: %q", got)
	}
}

func TestURLPathJoinWithNoParts(t *testing.T) {
	got, err := URLPathJoin()
	if err != nil {
		t.Fatalf("URLPathJoin returned error: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty URL, got %q", got)
	}
}
