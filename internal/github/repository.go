package github

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseRepository parses a repository reference such as:
//   - owner/name
//   - https://github.com/owner/name
//   - git@github.com:owner/name.git
//
// It returns the owner and repository name.
func ParseRepository(ref string) (string, string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("repository reference is empty")
	}

	// Handle git@github.com:owner/name(.git)
	if strings.HasPrefix(ref, "git@") {
		parts := strings.SplitN(ref, ":", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid repository reference: %s", ref)
		}
		ref = parts[1]
	}

	if strings.Contains(ref, "://") {
		u, err := url.Parse(ref)
		if err != nil {
			return "", "", fmt.Errorf("invalid repository url: %w", err)
		}
		return ParseRepositoryURL(u)
	}

	owner, name, err := parseOwnerAndName(ref)
	if err != nil {
		return "", "", err
	}
	return owner, name, ValidateRepository(owner, name)
}

// ParseRepositoryURL extracts the owner and name from a GitHub URL.
func ParseRepositoryURL(u *url.URL) (string, string, error) {
	if u == nil {
		return "", "", fmt.Errorf("url must not be nil")
	}

	path := strings.Trim(u.Path, "/")
	if path == "" {
		return "", "", fmt.Errorf("repository path is empty in url %s", u.String())
	}

	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("repository path must contain owner and name: %s", path)
	}

	owner := strings.TrimSpace(parts[0])
	name := strings.TrimSpace(parts[1])

	return owner, name, ValidateRepository(owner, name)
}

// CanonicalGitURL returns the canonical HTTPS URL for a GitHub repository.
func CanonicalGitURL(owner, name string) string {
	return fmt.Sprintf("https://github.com/%s/%s", owner, name)
}

// ValidateRepository ensures owner and name are both non-empty.
func ValidateRepository(owner, name string) error {
	if strings.TrimSpace(owner) == "" {
		return fmt.Errorf("repository owner is empty")
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("repository name is empty")
	}
	return nil
}

func parseOwnerAndName(ref string) (string, string, error) {
	ref = strings.TrimSuffix(ref, ".git")
	parts := strings.Split(ref, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("repository reference must contain owner and name: %s", ref)
	}

	owner := strings.TrimSpace(parts[len(parts)-2])
	name := strings.TrimSpace(parts[len(parts)-1])

	return owner, name, nil
}
