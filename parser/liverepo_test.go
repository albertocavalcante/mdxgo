package parser

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/albertocavalcante/mdxgo/syntax"
)

// TestLiveRepos clones real-world repos containing MDX/Markdown files and
// verifies the round-trip invariant against every file. This test is slow
// and requires network access, so it only runs when MDXGO_LIVE_REPOS=1.
//
// Run with:
//
//	MDXGO_LIVE_REPOS=1 go test -v -run TestLiveRepos -timeout 10m ./parser
func TestLiveRepos(t *testing.T) {
	if os.Getenv("MDXGO_LIVE_REPOS") != "1" {
		t.Skip("set MDXGO_LIVE_REPOS=1 to run live repo tests (requires network, slow)")
	}

	repos := []struct {
		url     string
		name    string // short name for temp dir
		mdx     bool   // parse in MDX mode
		shallow bool   // use shallow clone (depth=1)
		globs   []string
	}{
		{
			url:     "https://github.com/mintlify/docs.git",
			name:    "mintlify-docs",
			mdx:     true,
			shallow: true,
			globs:   []string{"**/*.mdx"},
		},
		{
			url:     "https://github.com/facebook/docusaurus.git",
			name:    "docusaurus",
			mdx:     true,
			shallow: true,
			globs:   []string{"website/docs/**/*.mdx", "website/docs/**/*.md"},
		},
		{
			url:     "https://github.com/mdx-js/mdx.git",
			name:    "mdx-js",
			mdx:     true,
			shallow: true,
			globs:   []string{"docs/**/*.mdx", "docs/**/*.md"},
		},
		{
			url:     "https://github.com/mintlify/starter.git",
			name:    "mintlify-starter",
			mdx:     true,
			shallow: true,
			globs:   []string{"**/*.mdx"},
		},
		{
			url:     "https://github.com/lightdash/mintlify-docs.git",
			name:    "lightdash-docs",
			mdx:     true,
			shallow: true,
			globs:   []string{"**/*.mdx"},
		},
		{
			url:     "https://github.com/commonmark/commonmark-spec.git",
			name:    "commonmark-spec",
			mdx:     false,
			shallow: true,
			globs:   []string{"**/*.md"},
		},
	}

	tmpRoot := t.TempDir()

	for _, repo := range repos {
		t.Run(repo.name, func(t *testing.T) {
			repoDir := filepath.Join(tmpRoot, repo.name)

			// Clone the repo.
			t.Logf("Cloning %s into %s...", repo.url, repoDir)
			args := []string{"clone", "--single-branch"}
			if repo.shallow {
				args = append(args, "--depth", "1")
			}
			args = append(args, repo.url, repoDir)

			cmd := exec.Command("git", args...)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				t.Fatalf("git clone failed: %v", err)
			}

			// Find all matching files.
			var files []string
			for _, glob := range repo.globs {
				matches, err := filepath.Glob(filepath.Join(repoDir, glob))
				if err != nil {
					t.Logf("glob %q error: %v", glob, err)
					continue
				}
				files = append(files, matches...)
			}

			// Also walk for nested globs that filepath.Glob misses.
			if len(files) == 0 {
				for _, glob := range repo.globs {
					ext := filepath.Ext(glob)
					if ext == "" {
						ext = ".mdx"
					}
					err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
						if err != nil {
							return nil
						}
						if !info.IsDir() && strings.HasSuffix(path, ext) {
							files = append(files, path)
						}
						return nil
					})
					if err != nil {
						t.Logf("walk error: %v", err)
					}
				}
			}

			if len(files) == 0 {
				t.Logf("WARNING: no matching files found in %s", repo.name)
				return
			}

			t.Logf("Found %d files in %s", len(files), repo.name)

			// Test each file.
			passed, failed := 0, 0
			opts := Options{MDX: repo.mdx}
			for _, fpath := range files {
				relPath, _ := filepath.Rel(repoDir, fpath)
				t.Run(relPath, func(t *testing.T) {
					data, err := os.ReadFile(fpath)
					if err != nil {
						t.Fatalf("read error: %v", err)
					}

					src := string(data)

					// Catch panics.
					var green *syntax.GreenNode
					func() {
						defer func() {
							if r := recover(); r != nil {
								t.Fatalf("PANIC on %s: %v", relPath, r)
							}
						}()
						green = Parse(data, opts)
					}()

					got := syntax.FullText(green)
					if got != src {
						failed++
						diverge := firstDivergence(src, got)
						t.Errorf("round-trip failed:\n"+
							"  file: %s\n"+
							"  input:  %d bytes\n"+
							"  output: %d bytes\n"+
							"  diverge at byte %d\n"+
							"  context: %q",
							relPath, len(src), len(got), diverge,
							safeSlice(src, max(0, diverge-30), min(len(src), diverge+30)),
						)
					} else {
						passed++
					}

					// Width must equal input length.
					if green.Width != len(data) {
						t.Errorf("width %d != len %d", green.Width, len(data))
					}
				})
			}

			t.Logf("Results for %s: %d/%d passed (%.1f%%)",
				repo.name, passed, passed+failed,
				float64(passed)/float64(passed+failed)*100)
		})
	}
}

// TestLiveRepoLocal tests against any MDX/MD files in a local directory.
// Set MDXGO_LOCAL_CORPUS=/path/to/dir to use.
//
// Example:
//
//	MDXGO_LOCAL_CORPUS=~/dev/ws/bazeldoc/docs go test -v -run TestLiveRepoLocal ./parser
func TestLiveRepoLocal(t *testing.T) {
	dir := os.Getenv("MDXGO_LOCAL_CORPUS")
	if dir == "" {
		t.Skip("set MDXGO_LOCAL_CORPUS=/path/to/dir to run local corpus tests")
	}

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("MDXGO_LOCAL_CORPUS=%q is not a valid directory: %v", dir, err)
	}

	var mdxFiles, mdFiles []string
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == ".next" {
				return filepath.SkipDir
			}
			return nil
		}
		switch filepath.Ext(path) {
		case ".mdx":
			mdxFiles = append(mdxFiles, path)
		case ".md":
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	t.Logf("Found %d .mdx files and %d .md files in %s", len(mdxFiles), len(mdFiles), dir)

	testFiles := func(files []string, opts Options) {
		passed, failed := 0, 0
		for _, fpath := range files {
			relPath, _ := filepath.Rel(dir, fpath)
			t.Run(relPath, func(t *testing.T) {
				data, err := os.ReadFile(fpath)
				if err != nil {
					t.Fatalf("read error: %v", err)
				}

				var green *syntax.GreenNode
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("PANIC: %v", r)
						}
					}()
					green = Parse(data, opts)
				}()

				got := syntax.FullText(green)
				src := string(data)
				if got != src {
					failed++
					diverge := firstDivergence(src, got)
					t.Errorf("round-trip failed: input=%d output=%d diverge=%d\n  context: %q",
						len(src), len(got), diverge,
						safeSlice(src, max(0, diverge-30), min(len(src), diverge+30)))
				} else {
					passed++
				}
			})
		}
		if passed+failed > 0 {
			t.Logf("%d/%d passed (%.1f%%)", passed, passed+failed,
				float64(passed)/float64(passed+failed)*100)
		}
	}

	if len(mdxFiles) > 0 {
		t.Run("mdx", func(t *testing.T) {
			testFiles(mdxFiles, Options{MDX: true})
		})
	}
	if len(mdFiles) > 0 {
		t.Run("md", func(t *testing.T) {
			testFiles(mdFiles, Options{MDX: false})
		})
	}
}

// TestLiveDoczelTestdata tests against doczel's own testdata MDX files
// if the doczel repo is available at the standard location.
func TestLiveDoczelTestdata(t *testing.T) {
	// Try common locations for the doczel repo.
	candidates := []string{
		"/Volumes/T9/dev/ws/doczel",
		os.ExpandEnv("$HOME/dev/ws/doczel"),
	}

	var doczelDir string
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.IsDir() {
			doczelDir = c
			break
		}
	}
	if doczelDir == "" {
		t.Skip("doczel repo not found at standard locations")
	}

	var mdxFiles []string
	err := filepath.Walk(doczelDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if base == ".git" || base == "node_modules" || base == "vendor" || base == "ui" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) == ".mdx" {
			mdxFiles = append(mdxFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk error: %v", err)
	}

	if len(mdxFiles) == 0 {
		t.Skip("no .mdx files found in doczel")
	}

	t.Logf("Testing %d .mdx files from doczel testdata", len(mdxFiles))

	passed, failed := 0, 0
	for _, fpath := range mdxFiles {
		relPath, _ := filepath.Rel(doczelDir, fpath)
		t.Run(relPath, func(t *testing.T) {
			data, err := os.ReadFile(fpath)
			if err != nil {
				t.Fatalf("read error: %v", err)
			}

			var green *syntax.GreenNode
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("PANIC: %v", r)
					}
				}()
				green = Parse(data, Options{MDX: true})
			}()

			got := syntax.FullText(green)
			src := string(data)
			if got != src {
				failed++
				diverge := firstDivergence(src, got)
				t.Errorf("round-trip failed: input=%d output=%d diverge=%d",
					len(src), len(got), diverge)
			} else {
				passed++
			}
		})
	}

	if passed+failed > 0 {
		pct := float64(passed) / float64(passed+failed) * 100
		t.Logf("doczel testdata: %d/%d passed (%.1f%%)", passed, passed+failed, pct)
		if failed > 0 {
			t.Logf("NOTE: %d failures expected — doczel MDX files may use inline constructs "+
				"that haven't been tested in block-only mode", failed)
		}
	}
}

// TestLiveSingleFile tests a single file path for debugging.
// Set MDXGO_TEST_FILE=/path/to/file.mdx to use.
func TestLiveSingleFile(t *testing.T) {
	fpath := os.Getenv("MDXGO_TEST_FILE")
	if fpath == "" {
		t.Skip("set MDXGO_TEST_FILE=/path/to/file to test a single file")
	}

	data, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	mdx := strings.HasSuffix(fpath, ".mdx")
	green := Parse(data, Options{MDX: mdx})

	got := syntax.FullText(green)
	src := string(data)
	if got != src {
		diverge := firstDivergence(src, got)
		t.Errorf("round-trip failed at byte %d", diverge)

		// Show context around divergence.
		lo := max(0, diverge-50)
		hi := min(len(src), diverge+50)
		t.Logf("input context:  %q", src[lo:hi])
		lo2 := max(0, diverge-50)
		hi2 := min(len(got), diverge+50)
		t.Logf("output context: %q", got[lo2:hi2])
	} else {
		t.Logf("Round-trip OK: %d bytes", len(src))
	}

	// Dump tree structure.
	t.Logf("Tree dump:\n%s", syntax.DebugDump(green))
	t.Logf("Root width: %d, input length: %d", green.Width, len(data))
	t.Logf("Children: %d", len(green.Children))
	for i, child := range green.Children {
		t.Logf("  child[%d]: kind=%v width=%d", i, child.Kind(), child.Width())
	}

	fmt.Fprintf(os.Stderr, "\nTree:\n%s\n", syntax.DebugDump(green))
}
