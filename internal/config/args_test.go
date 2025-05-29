package config

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/edulinq/autograder/internal/util"
)

func TestConfigArgs(test *testing.T) {
	defer Reset()

	tempDir := setupConfigTempDir(test)
	defer util.RemoveDirent(tempDir)

	testCases := []struct {
		baseDir        string
		cwd            string
		args           ConfigArgs
		errorSubstring string
		expected       map[string]any
	}{
		// Empty config.
		{
			filepath.Join(tempDir, "empty"),
			tempDir,
			ConfigArgs{},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
			},
		},

		// Base dirs cannot be set in config files.
		{
			filepath.Join(tempDir, "bad-base-dir"),
			tempDir,
			ConfigArgs{},
			"Cannot set key 'dirs.base' in config files, use the command-line.",
			nil,
		},

		// config and secrets file without overriding.
		{
			filepath.Join(tempDir, "dual-different"),
			tempDir,
			ConfigArgs{},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "dual-different"),
				"a":         "A",
				"b":         "B",
			},
		},

		// config and secrets file without overriding.
		{
			filepath.Join(tempDir, "dual-override"),
			tempDir,
			ConfigArgs{},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "dual-override"),
				"a":         "secret",
			},
		},

		// Config on cmd.
		{
			filepath.Join(tempDir, "empty"),
			tempDir,
			ConfigArgs{
				Config: map[string]string{
					"a": "A",
				},
			},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
				"a":         "A",
			},
		},

		// Load config and path on cmd.
		{
			filepath.Join(tempDir, "empty"),
			tempDir,
			ConfigArgs{
				Config: map[string]string{
					"c": "C",
				},
				ConfigPath: []string{
					getConfigPath(tempDir, "dual-different", "config.json"),
				},
			},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
				"a":         "A",
				"c":         "C",
			},
		},

		// Load config and path on cmd, with overriding.
		{
			filepath.Join(tempDir, "empty"),
			tempDir,
			ConfigArgs{
				Config: map[string]string{
					"a": "ZZZ",
				},
				ConfigPath: []string{
					getConfigPath(tempDir, "dual-different", "config.json"),
				},
			},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
				"a":         "ZZZ",
			},
		},

		// config file, secrets file, cmd, config path; with override.
		{
			filepath.Join(tempDir, "dual-override"),
			tempDir,
			ConfigArgs{
				Config: map[string]string{
					"a": "ZZZ",
				},
				ConfigPath: []string{
					getConfigPath(tempDir, "dual-different", "config.json"),
				},
			},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "dual-override"),
				"a":         "ZZZ",
			},
		},

		// cwd config and cwd secrets file without overriding.
		{
			filepath.Join(tempDir, "empty"),
			getConfigDir(tempDir, "dual-different"),
			ConfigArgs{},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
				"a":         "A",
				"b":         "B",
			},
		},

		// cwd config and cwd secrets file with overriding.
		{
			filepath.Join(tempDir, "empty"),
			getConfigDir(tempDir, "dual-override"),
			ConfigArgs{},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "empty"),
				"a":         "secret",
			},
		},

		// sys config and secrets, cwd config and secrets, cmd, config path; with override.
		{
			filepath.Join(tempDir, "dual-override"),
			getConfigDir(tempDir, "dual-override"),
			ConfigArgs{
				Config: map[string]string{
					"a": "ZZZ",
				},
				ConfigPath: []string{
					getConfigPath(tempDir, "dual-different", "config.json"),
				},
			},
			"",
			map[string]any{
				"dirs.base": filepath.Join(tempDir, "dual-override"),
				"a":         "ZZZ",
			},
		},
	}

	for i, testCase := range testCases {
		Reset()
		BASE_DIR.Set(testCase.baseDir)

		err := HandleConfigArgsFull(testCase.args, testCase.cwd, true)
		if err != nil {
			if testCase.errorSubstring != "" {
				if !strings.Contains(err.Error(), testCase.errorSubstring) {
					test.Errorf("Case %d: Did not get expected error output. Expected Substring '%s', Actual Error: '%v'.", i, testCase.errorSubstring, err)
				}
			} else {
				test.Errorf("Case %d: Failed to handle args: '%v'.", i, err)
			}

			continue
		} else if testCase.errorSubstring != "" {
			test.Errorf("Case %d: Did not get expected error on '%s'.", i, testCase.errorSubstring)
			continue
		}

		if !reflect.DeepEqual(testCase.expected, configValues) {
			test.Errorf("Case %d: Config values not as expected. Expected: '%v', Actual: '%v'.", i, testCase.expected, configValues)
			continue
		}
	}
}

func setupConfigTempDir(test *testing.T) string {
	tempDir, err := util.MkDirTemp("autograder-config-")
	if err != nil {
		test.Fatalf("Failed to create temp dir: '%v'.", err)
	}

	contents := []struct {
		text string
		path string
	}{
		{
			text: `{}`,
			path: getConfigPath(tempDir, "empty", "config.json"),
		},

		{
			text: `{"dirs.base": "/dev/null"}`,
			path: getConfigPath(tempDir, "bad-base-dir", "config.json"),
		},

		{
			text: `{"a": "A"}`,
			path: getConfigPath(tempDir, "dual-different", "config.json"),
		},
		{
			text: `{"b": "B"}`,
			path: getConfigPath(tempDir, "dual-different", "secrets.json"),
		},

		{
			text: `{"a": "config"}`,
			path: getConfigPath(tempDir, "dual-override", "config.json"),
		},
		{
			text: `{"a": "secret"}`,
			path: getConfigPath(tempDir, "dual-override", "secrets.json"),
		},
	}

	for _, content := range contents {
		err = util.MkDir(filepath.Dir(content.path))
		if err != nil {
			test.Fatalf("Failed to make temp subdir: '%v'.", err)
		}

		err = util.WriteFile(content.text, content.path)
		if err != nil {
			test.Fatalf("Failed to write contents to '%s': '%v'.", content.path, err)
		}
	}

	return tempDir
}

func getConfigDir(tempDir string, basename string) string {
	// Hard-coded specifically for testing.
	return filepath.Join(tempDir, basename, NAME.Get(), CONFIG_DIRNAME)
}

func getConfigPath(tempDir string, basename string, filename string) string {
	return filepath.Join(getConfigDir(tempDir, basename), filename)
}
