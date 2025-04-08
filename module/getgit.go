// SPDX-FileCopyrightText: 2025 Masa Cento
// SPDX-License-Identifier: MPL-2.0

package module

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

// GetByGit fetches a module from a git repository and stores it to the keyvalue store.
func GetByGit(ctx context.Context, kv jetstream.KeyValue, module string) (*Module, error) {
	slog.Info("jumon get by git", "module", module)
	repo, path, err := getVCSPath(module)
	if err != nil {
		return nil, fmt.Errorf("invalid module name: %w", err)
	}
	tempDir, err := os.MkdirTemp("", "jumon-git-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	checkoutDir, err := sparseCheckout(repo, path, tempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to git sparse checkout: %w", err)
	}
	slog.Debug("checkout directory", "dir", checkoutDir)
	return GetByDir(ctx, kv, checkoutDir)
}

// getVCSPath extracts repository URL and path from a module path.
// Only GitHub format (github.com/user/repo/path) is fully supported.
func getVCSPath(module string) (repo string, path string, err error) {
	for _, vcsPath := range vcsPaths {
		m := vcsPath.regexp.FindStringSubmatch(module)
		if len(m) == 0 {
			continue
		}
		match := map[string]string{}
		for i, name := range vcsPath.regexp.SubexpNames() {
			if name != "" && match[name] == "" {
				match[name] = m[i]
			}
		}
		repo = expand(match, vcsPath.repo)
		path = expand(match, vcsPath.dir)
		return
	}

	return "", "", fmt.Errorf("invalid module path: %s", module)
}

type vcsPath struct {
	pathPrefix string
	regexp     *regexp.Regexp
	repo       string
	dir        string
}

var vcsPaths = []*vcsPath{
	// GitHub
	{
		pathPrefix: "github.com",
		regexp:     regexp.MustCompile(`^(?P<root>github\.com/[\w.\-]+/[\w.\-]+)(?:/(?P<dir>[\w.\-]+(?:/[\w.\-]+)*))?$`),
		repo:       "https://{root}",
		dir:        "{dir}",
	},
	// Other
	{
		regexp: regexp.MustCompile(`^(?P<root>([a-z0-9.\-]+\.)+[a-z0-9.\-]+(:[0-9]+)?(/~?[\w.\-]+)+?(/~?[\w.\-]+)+?)(?:/(?P<dir>[\w.\-]+(?:/[\w.\-]+)*))?$`),
		repo:   "https://{root}",
		dir:    "{dir}",
	},
}

// sparseCheckout performs a git sparse-checkout of the specified path.
// Returns the path to the checked out directory.
func sparseCheckout(repoURL, repoPath, destDir string) (string, error) {
	err := gitCommand("", "clone", "--depth", "1", "--filter=blob:none", "--no-checkout", repoURL, destDir)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}

	err = gitCommand(destDir, "sparse-checkout", "init", "--cone")
	if err != nil {
		return "", fmt.Errorf("failed to initialize sparse-checkout: %w", err)
	}

	err = gitCommand(destDir, "sparse-checkout", "set", repoPath)
	if err != nil {
		return "", fmt.Errorf("failed to set sparse-checkout: %w", err)
	}

	err = gitCommand(destDir, "checkout")
	if err != nil {
		return "", fmt.Errorf("failed to checkout: %w", err)
	}
	return filepath.Join(destDir, repoPath), nil
}

// gitCommand executes a git command with the given arguments.
func gitCommand(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// expand replaces placeholders in the format {key} with values from the map.
// e.g. match = {"key": "value"} -> s = "https://{key}" -> result = "https://value"
func expand(match map[string]string, s string) string {
	result := s
	for k, v := range match {
		result = strings.ReplaceAll(result, "{"+k+"}", v)
	}
	return result
}
