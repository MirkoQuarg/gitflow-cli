/*
SPDX-FileCopyrightText: 2026 Mercedes-Benz Tech Innovation GmbH
SPDX-License-Identifier: MIT
*/

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Git reports conflicting files relative to the repository root, while a
// plugin names its version file relative to the project path. These are the
// cases that translation has to get right for a project path that is not the
// repository root itself.
func TestRelativeToPrefix(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		fileName string
		expected string
	}{
		{
			name:     "project path is the repository root",
			prefix:   "",
			fileName: "pom.xml",
			expected: "pom.xml",
		},
		{
			name:     "file inside the project path",
			prefix:   "agent/",
			fileName: "agent/pyproject.toml",
			expected: "pyproject.toml",
		},
		{
			name:     "file deeper inside the project path",
			prefix:   "agent/",
			fileName: "agent/sub/version.txt",
			expected: "sub/version.txt",
		},
		{
			name:     "nested project path",
			prefix:   "services/agent/",
			fileName: "services/agent/pyproject.toml",
			expected: "pyproject.toml",
		},
		{
			name:     "file outside the project path climbs out of it",
			prefix:   "agent/",
			fileName: "README.md",
			expected: "../README.md",
		},
		{
			name:     "file outside a nested project path climbs out twice",
			prefix:   "services/agent/",
			fileName: "version.txt",
			expected: "../../version.txt",
		},
		{
			// the trailing slash is what keeps this from being mistaken for a
			// file inside the project path
			name:     "a sibling whose name starts the same is not inside",
			prefix:   "agent/",
			fileName: "agentic/pyproject.toml",
			expected: "../agentic/pyproject.toml",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, relativeToPrefix(test.prefix, test.fileName))
		})
	}
}
