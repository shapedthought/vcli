package resources_test

import (
	"testing"

	"github.com/shapedthought/owlctl/resources"
)

func TestIsMixinKind(t *testing.T) {
	tests := []struct {
		kind string
		want bool
	}{
		{resources.KindProfile, true},
		{resources.KindOverlay, true},
		{resources.KindVBRJob, false},
		{resources.KindVBRRepository, false},
		{resources.KindVBRSOBR, false},
		{resources.KindVBRScaleOutRepository, false},
		{resources.KindVBREncryptionPassword, false},
		{resources.KindVBRKmsServer, false},
		{"Unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := resources.IsMixinKind(tt.kind)
			if got != tt.want {
				t.Errorf("IsMixinKind(%q) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}

func TestIsResourceKind(t *testing.T) {
	tests := []struct {
		kind string
		want bool
	}{
		{resources.KindVBRJob, true},
		{resources.KindVBRRepository, true},
		{resources.KindVBRSOBR, true},
		{resources.KindVBRScaleOutRepository, true},
		{resources.KindVBREncryptionPassword, true},
		{resources.KindVBRKmsServer, true},
		{resources.KindProfile, false},
		{resources.KindOverlay, false},
		{"Unknown", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			got := resources.IsResourceKind(tt.kind)
			if got != tt.want {
				t.Errorf("IsResourceKind(%q) = %v, want %v", tt.kind, got, tt.want)
			}
		})
	}
}
