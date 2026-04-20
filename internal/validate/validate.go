package validate

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// safeNameRegex allows only alphanumeric, hyphens, and underscores.
var safeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

// ProjectName validates a project name is safe for use as a directory name,
// Docker network name, and container label value.
func ProjectName(name string) error {
	if !safeNameRegex.MatchString(name) {
		return fmt.Errorf("project name must be 1-64 alphanumeric characters, hyphens, or underscores, starting with alphanumeric")
	}
	return nil
}

// GitURL validates that a URL is a safe git remote.
// Only allows https:// and git:// schemes.
func GitURL(rawURL string) error {
	if rawURL == "" {
		return nil // optional field
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	switch parsed.Scheme {
	case "https", "git", "file":
		// OK
	case "":
		// Absolute path or relative path
		if parsed.Path == "" {
			return fmt.Errorf("invalid empty path")
		}
		// Basic check for "garbage"
		if strings.ContainsAny(rawURL, " !@#$^&*()<>?") && !strings.Contains(rawURL, "://") {
			return fmt.Errorf("invalid characters in local path")
		}
	default:
		// Check for Windows drive letters (e.g. C:)
		if len(parsed.Scheme) == 1 && ((parsed.Scheme[0] >= 'a' && parsed.Scheme[0] <= 'z') || (parsed.Scheme[0] >= 'A' && parsed.Scheme[0] <= 'Z')) {
			// OK
		} else {
			return fmt.Errorf("unsupported URL scheme %q", parsed.Scheme)
		}
	}

	if parsed.Host == "" && (parsed.Scheme == "https" || parsed.Scheme == "git") {
		return fmt.Errorf("remote URL must have a host")
	}

	// Block embedded credentials
	if parsed.User != nil {
		return fmt.Errorf("URLs with embedded credentials are not allowed")
	}

	return nil
}

// WorkspacePath ensures the resolved path stays inside the base workspace directory.
func WorkspacePath(basePath, projectName string) (string, error) {
	full := filepath.Join(basePath, projectName)
	abs, err := filepath.Abs(full)
	if err != nil {
		return "", fmt.Errorf("cannot resolve path: %w", err)
	}

	baseAbs, err := filepath.Abs(basePath)
	if err != nil {
		return "", fmt.Errorf("cannot resolve base path: %w", err)
	}

	// Ensure the resolved path is inside the base
	if !strings.HasPrefix(abs, baseAbs+string(filepath.Separator)) && abs != baseAbs {
		return "", fmt.Errorf("path traversal detected: resolved path %q escapes workspace %q", abs, baseAbs)
	}

	return abs, nil
}

// ContainerID validates a Docker container ID or name.
// Docker IDs are hex strings; names are alphanumeric with some punctuation.
var containerIDRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.\-/:]{0,127}$`)

func ContainerID(id string) error {
	if !containerIDRegex.MatchString(id) {
		return fmt.Errorf("invalid container identifier")
	}
	return nil
}
