package stringx

import "testing"

func TestCamelToSnakeHandlesInitialisms(t *testing.T) {
	tests := map[string]string{
		"UserName":        "user_name",
		"HTTPServer":      "http_server",
		"UserID":          "user_id",
		"UserIDAndTeamID": "user_id_and_team_id",
		"APIKey":          "api_key",
		"BaseURL":         "base_url",
		"userName":        "user_name",
		"user_name":       "user_name",
		"":                "",
	}

	for input, expected := range tests {
		t.Run(input, func(t *testing.T) {
			if got := CamelToSnake(input); got != expected {
				t.Fatalf("CamelToSnake(%q) = %q, want %q", input, got, expected)
			}
		})
	}
}
