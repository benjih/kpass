package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func makeDirs(t *testing.T, names ...string) []string {
	t.Helper()
	base := t.TempDir()
	var paths []string
	for _, name := range names {
		p := filepath.Join(base, name)
		if err := os.Mkdir(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", p, err)
		}
		paths = append(paths, p)
	}
	return paths
}

func TestAugmentEnvPathFiltered(t *testing.T) {
	t.Run("prepends new paths to empty env var", func(t *testing.T) {
		dirs := makeDirs(t, "a", "b")
		key := "TEST_AUGMENT_EMPTY"
		t.Setenv(key, "")

		AugmentEnvPathFiltered(key, nil, dirs...)

		got := os.Getenv(key)
		if got != strings.Join(dirs, ":") {
			t.Errorf("got %q, want %q", got, strings.Join(dirs, ":"))
		}
	})

	t.Run("new paths prepended before existing", func(t *testing.T) {
		dirs := makeDirs(t, "new", "existing")
		newDir, existingDir := dirs[0], dirs[1]
		key := "TEST_AUGMENT_ORDER"
		t.Setenv(key, existingDir)

		AugmentEnvPathFiltered(key, nil, newDir)

		got := os.Getenv(key)
		if got != newDir+":"+existingDir {
			t.Errorf("got %q, want %q", got, newDir+":"+existingDir)
		}
	})

	t.Run("duplicates not added twice", func(t *testing.T) {
		dirs := makeDirs(t, "dup")
		key := "TEST_AUGMENT_DUP"
		t.Setenv(key, dirs[0])

		AugmentEnvPathFiltered(key, nil, dirs[0])

		got := os.Getenv(key)
		if got != dirs[0] {
			t.Errorf("got %q, want single entry %q", got, dirs[0])
		}
	})

	t.Run("non-existent paths are skipped", func(t *testing.T) {
		key := "TEST_AUGMENT_NONEXIST"
		t.Setenv(key, "")

		AugmentEnvPathFiltered(key, nil, "/does/not/exist/at/all")

		got := os.Getenv(key)
		if got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("skipSubstrings filters matching paths", func(t *testing.T) {
		dirs := makeDirs(t, "keep", "skip-me")
		keepDir, skipDir := dirs[0], dirs[1]
		key := "TEST_AUGMENT_SKIP"
		t.Setenv(key, "")

		AugmentEnvPathFiltered(key, []string{"skip-me"}, keepDir, skipDir)

		got := os.Getenv(key)
		if got != keepDir {
			t.Errorf("got %q, want only %q", got, keepDir)
		}
	})

	t.Run("skipSubstrings also filters existing env entries", func(t *testing.T) {
		dirs := makeDirs(t, "keep2", "bad-prefix")
		keepDir, badDir := dirs[0], dirs[1]
		key := "TEST_AUGMENT_SKIP_EXISTING"
		t.Setenv(key, badDir)

		AugmentEnvPathFiltered(key, []string{"bad-prefix"}, keepDir)

		got := os.Getenv(key)
		if got != keepDir {
			t.Errorf("got %q, want only %q", got, keepDir)
		}
	})

	t.Run("empty string path is skipped", func(t *testing.T) {
		dirs := makeDirs(t, "real")
		key := "TEST_AUGMENT_EMPTY_PATH"
		t.Setenv(key, "")

		AugmentEnvPathFiltered(key, nil, "", dirs[0])

		got := os.Getenv(key)
		if got != dirs[0] {
			t.Errorf("got %q, want %q", got, dirs[0])
		}
	})
}

func TestAugmentEnvPath(t *testing.T) {
	dirs := makeDirs(t, "smoke")
	key := "TEST_AUGMENT_SMOKE"
	t.Setenv(key, "")

	AugmentEnvPath(key, dirs[0])

	got := os.Getenv(key)
	if got != dirs[0] {
		t.Errorf("AugmentEnvPath: got %q, want %q", got, dirs[0])
	}
}
