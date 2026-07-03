package auth

import "testing"

func TestParseBearer(t *testing.T) {
	token, err := ParseBearer("Bearer abc123")
	if err != nil {
		t.Fatal(err)
	}
	if token != "abc123" {
		t.Fatalf("token = %q", token)
	}
}