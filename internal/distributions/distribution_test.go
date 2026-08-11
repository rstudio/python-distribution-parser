package distributions

import (
	"strings"
	"testing"
)

func TestHeaderAttrsVersions(t *testing.T) {
	expectedVersions := []string{"1.0", "1.1", "1.2", "2.0", "2.1", "2.2", "2.3", "2.4", "2.5"}

	for _, version := range expectedVersions {
		t.Run("version_"+version, func(t *testing.T) {
			attrs, exists := HeaderAttrs[version]
			if !exists {
				t.Errorf("metadata version %s not found in HeaderAttrs map", version)
				return
			}
			if len(attrs) == 0 {
				t.Errorf("metadata version %s has no header attributes", version)
			}
		})
	}
}

func TestGetHeaderAttrs(t *testing.T) {
	tests := []struct {
		version     string
		expectError bool
	}{
		{"1.0", false},
		{"1.1", false},
		{"1.2", false},
		{"2.0", false},
		{"2.1", false},
		{"2.2", false},
		{"2.3", false},
		{"2.4", false},
		{"2.5", false},
		{"9.9", true},
	}

	for _, tt := range tests {
		t.Run("version_"+tt.version, func(t *testing.T) {
			bd := &BaseDistribution{MetadataVersion: tt.version}
			attrs, err := bd.GetHeaderAttrs()

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for version %s, got nil", tt.version)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for version %s: %v", tt.version, err)
				}
				if len(attrs) == 0 {
					t.Errorf("expected non-empty attrs for version %s", tt.version)
				}
			}
		})
	}
}

func TestHeaderAttrs2_5HasLicenseExpression(t *testing.T) {
	attrs := HeaderAttrs2_5
	found := false
	for _, attr := range attrs {
		if attr.HeaderName == "License-Expression" {
			found = true
			break
		}
	}
	if !found {
		t.Error("HeaderAttrs2_5 should include License-Expression header")
	}
}

func TestParseMetadataVersion2_5(t *testing.T) {
	metadata := `Metadata-Version: 2.5
Name: test-package
Version: 1.0.0
Summary: A test package
License-Expression: MIT

`
	bd := &BaseDistribution{}
	err := bd.Parse([]byte(metadata))
	if err != nil {
		t.Fatalf("failed to parse metadata version 2.5: %v", err)
	}

	if bd.MetadataVersion != "2.5" {
		t.Errorf("expected metadata version 2.5, got %s", bd.MetadataVersion)
	}
	if bd.Name != "test-package" {
		t.Errorf("expected name test-package, got %s", bd.Name)
	}
	if bd.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", bd.Version)
	}
	if bd.LicenseExpression != "MIT" {
		t.Errorf("expected license expression MIT, got %s", bd.LicenseExpression)
	}
}

func TestParseMetadataWithDescription(t *testing.T) {
	metadata := `Metadata-Version: 2.5
Name: test-package
Version: 1.0.0

This is the description body.
It can have multiple lines.`

	bd := &BaseDistribution{}
	err := bd.Parse([]byte(metadata))
	if err != nil {
		t.Fatalf("failed to parse metadata: %v", err)
	}

	if !strings.Contains(bd.Description, "This is the description body") {
		t.Errorf("description not parsed correctly: %s", bd.Description)
	}
}

func TestHeaderAttrsInheritance(t *testing.T) {
	hasAttr := func(attrs []HeaderAttr, name string) bool {
		for _, a := range attrs {
			if a.HeaderName == name {
				return true
			}
		}
		return false
	}

	t.Run("2.1_has_provides_extra", func(t *testing.T) {
		if !hasAttr(HeaderAttrs2_1, "Provides-Extra") {
			t.Error("2.1 should have Provides-Extra")
		}
	})

	t.Run("2.2_has_dynamic", func(t *testing.T) {
		if !hasAttr(HeaderAttrs2_2, "Dynamic") {
			t.Error("2.2 should have Dynamic")
		}
	})

	t.Run("2.4_has_license_expression", func(t *testing.T) {
		if !hasAttr(HeaderAttrs2_4, "License-Expression") {
			t.Error("2.4 should have License-Expression")
		}
	})

	t.Run("2.5_inherits_from_2.4", func(t *testing.T) {
		if !hasAttr(HeaderAttrs2_5, "License-Expression") {
			t.Error("2.5 should inherit License-Expression from 2.4")
		}
		if !hasAttr(HeaderAttrs2_5, "Dynamic") {
			t.Error("2.5 should inherit Dynamic from 2.2")
		}
		if !hasAttr(HeaderAttrs2_5, "Provides-Extra") {
			t.Error("2.5 should inherit Provides-Extra from 2.1")
		}
	})
}
