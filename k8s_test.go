package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractBearer(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   string
	}{
		{
			name:   "valid bearer token",
			header: "Bearer i-am-a-bearer-token",
			want:   "i-am-a-bearer-token",
		},
		{
			name:   "missing prefix",
			header: "i-might-be-a-bearer-token-but-i-have-no-Bearer-before-me",
			want:   "",
		},
		{
			name:   "empty header",
			header: "",
			want:   "",
		},
		{
			name:   "prefix only",
			header: "Bearer ",
			want:   "",
		},
		{
			name:   "case sensitive prefix",
			header: "bearer i-could-be-a-bearer-token-but-the-guy-before-me-ruined-it",
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBearer(tt.header)
			assert.Equal(t, got, tt.want)
		})
	}
}
