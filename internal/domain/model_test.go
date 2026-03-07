package domain

import "testing"

func TestTag_IsInternal(t *testing.T) {
	tests := []struct {
		name string
		tag  Tag
		want bool
	}{
		{"internal tag", Tag{Name: "#internal"}, true},
		{"public tag", Tag{Name: "Go"}, false},
		{"empty name", Tag{Name: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tag.IsInternal(); got != tt.want {
				t.Errorf("IsInternal() = %v, want %v", got, tt.want)
			}
		})
	}
}
