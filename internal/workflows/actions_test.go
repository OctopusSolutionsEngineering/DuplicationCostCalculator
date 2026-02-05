package workflows

import "testing"

func TestGetActionIdAndVersion(t *testing.T) {
	tests := []struct {
		name            string
		uses            string
		expectedId      string
		expectedVersion string
	}{
		{
			name:            "action with version",
			uses:            "actions/checkout@v4",
			expectedId:      "actions/checkout",
			expectedVersion: "v4",
		},
		{
			name:            "action with semantic version",
			uses:            "actions/setup-node@v3.5.1",
			expectedId:      "actions/setup-node",
			expectedVersion: "v3.5.1",
		},
		{
			name:            "action with SHA version",
			uses:            "actions/checkout@8e5e7e5ab8b370d6c329ec480221332ada57f0ab",
			expectedId:      "actions/checkout",
			expectedVersion: "8e5e7e5ab8b370d6c329ec480221332ada57f0ab",
		},
		{
			name:            "action without version",
			uses:            "actions/checkout",
			expectedId:      "actions/checkout",
			expectedVersion: "latest",
		},
		{
			name:            "action with v prefix",
			uses:            "docker/build-push-action@v5",
			expectedId:      "docker/build-push-action",
			expectedVersion: "v5",
		},
		{
			name:            "action with major.minor.patch version",
			uses:            "hashicorp/setup-terraform@v2.0.3",
			expectedId:      "hashicorp/setup-terraform",
			expectedVersion: "v2.0.3",
		},
		{
			name:            "action with multiple @ symbols (edge case)",
			uses:            "org/action@v1@extra",
			expectedId:      "org/action",
			expectedVersion: "v1@extra",
		},
		{
			name:            "empty string",
			uses:            "",
			expectedId:      "",
			expectedVersion: "latest",
		},
		{
			name:            "action with numeric version",
			uses:            "actions/cache@3",
			expectedId:      "actions/cache",
			expectedVersion: "3",
		},
		{
			name:            "action with prerelease version",
			uses:            "actions/setup-go@v4.1.0-rc.1",
			expectedId:      "actions/setup-go",
			expectedVersion: "v4.1.0-rc.1",
		},
		{
			name:            "action with build metadata",
			uses:            "actions/cache@v3.0.0+build.123",
			expectedId:      "actions/cache",
			expectedVersion: "v3.0.0+build.123",
		},
		{
			name:            "action with branch name",
			uses:            "actions/checkout@main",
			expectedId:      "actions/checkout",
			expectedVersion: "main",
		},
		{
			name:            "action with tag name",
			uses:            "docker/login-action@v2.1.0",
			expectedId:      "docker/login-action",
			expectedVersion: "v2.1.0",
		},
		{
			name:            "only @ symbol",
			uses:            "@",
			expectedId:      "",
			expectedVersion: "",
		},
		{
			name:            "@ at the end",
			uses:            "actions/checkout@",
			expectedId:      "actions/checkout",
			expectedVersion: "",
		},
		{
			name:            "organization/repo/action format",
			uses:            "octocat/hello-world-docker-action@v1.2.3",
			expectedId:      "octocat/hello-world-docker-action",
			expectedVersion: "v1.2.3",
		},
		{
			name:            "local action path",
			uses:            "./.github/actions/my-action",
			expectedId:      "./.github/actions/my-action",
			expectedVersion: "latest",
		},
		{
			name:            "action with just latest",
			uses:            "actions/cache@latest",
			expectedId:      "actions/cache",
			expectedVersion: "latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, version := GetActionIdAndVersion(tt.uses)
			if id != tt.expectedId {
				t.Errorf("GetActionIdAndVersion(%q) id = %q, expected %q", tt.uses, id, tt.expectedId)
			}
			if version != tt.expectedVersion {
				t.Errorf("GetActionIdAndVersion(%q) version = %q, expected %q", tt.uses, version, tt.expectedVersion)
			}
		})
	}
}

func TestGetActionIdAndVersionConsistency(t *testing.T) {
	// Test that calling the function multiple times returns consistent results
	uses := "actions/checkout@v4"

	id1, version1 := GetActionIdAndVersion(uses)
	id2, version2 := GetActionIdAndVersion(uses)
	id3, version3 := GetActionIdAndVersion(uses)

	if id1 != id2 || id2 != id3 {
		t.Errorf("GetActionIdAndVersion(%q) returned inconsistent IDs: %q, %q, %q", uses, id1, id2, id3)
	}

	if version1 != version2 || version2 != version3 {
		t.Errorf("GetActionIdAndVersion(%q) returned inconsistent versions: %q, %q, %q", uses, version1, version2, version3)
	}
}
