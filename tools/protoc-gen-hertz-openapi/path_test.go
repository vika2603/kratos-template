package main

import "testing"

func TestConvertPath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/api/v1/users", "/api/v1/users"},
		{"/api/v1/users/:id", "/api/v1/users/{id}"},
		{"/api/v1/users/:id/posts/:postId", "/api/v1/users/{id}/posts/{postId}"},
		{"/*filepath", "/{filepath}"},
		{"/static/*filepath", "/static/{filepath}"},
		{"/", "/"},
		{"", ""},
	}
	for _, tt := range tests {
		got := ConvertPath(tt.input)
		if got != tt.want {
			t.Errorf("ConvertPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
