package take

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"os/exec"
)

type testCase struct {
	name        string
	opts        Options
	setup       func(t *testing.T)
	cleanup     func() error
	wantResult  Result
	wantErr     error
	checkResult func(t *testing.T, got Result)
}

func createMockTarball(t *testing.T) string {
	dir, err := os.MkdirTemp("", "take-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test directory with a file
	testDir := filepath.Join(dir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create tarball
	tarPath := filepath.Join(dir, "test.tar.gz")
	cmd := exec.Command("tar", "czf", tarPath, "-C", dir, "testdir")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create tarball: %v\nOutput: %s", err, out)
	}

	return tarPath
}

func createMockZip(t *testing.T) string {
	dir, err := os.MkdirTemp("", "take-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a test directory with a file
	testDir := filepath.Join(dir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	testFile := filepath.Join(testDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create zip
	zipPath := filepath.Join(dir, "test.zip")
	cmd := exec.Command("zip", "-r", zipPath, "testdir")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create zip: %v\nOutput: %s", err, out)
	}

	return zipPath
}

func TestTake(t *testing.T) {
	// Create mock archives
	tarPath := createMockTarball(t)
	zipPath := createMockZip(t)
	defer os.RemoveAll(filepath.Dir(tarPath))
	defer os.RemoveAll(filepath.Dir(zipPath))

	// Create test server for archive downloads
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test.tar.gz":
			content, err := os.ReadFile(tarPath)
			if err != nil {
				t.Fatalf("Failed to read tarball: %v", err)
			}
			w.Write(content)
		case "/test.zip":
			content, err := os.ReadFile(zipPath)
			if err != nil {
				t.Fatalf("Failed to read zip: %v", err)
			}
			w.Write(content)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temporary directory
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalWd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Helper function to create paths relative to temp dir
	tmpPath := func(path string) string {
		return filepath.Join(tmpDir, path)
	}

	tests := []testCase{
		{
			name: "create new directory",
			opts: Options{
				Path: tmpPath("newdir"),
			},
			checkResult: func(t *testing.T, got Result) {
				if !got.WasCreated {
					t.Error("Expected directory to be created")
				}
				if _, err := os.Stat(got.FinalPath); os.IsNotExist(err) {
					t.Error("Directory was not created")
				}
			},
		},
		{
			name: "create nested directories",
			opts: Options{
				Path: tmpPath("parent/child/grandchild"),
			},
			checkResult: func(t *testing.T, got Result) {
				if !got.WasCreated {
					t.Error("Expected directories to be created")
				}
				if _, err := os.Stat(got.FinalPath); os.IsNotExist(err) {
					t.Error("Directories were not created")
				}
			},
		},
		{
			name: "handle existing directory",
			setup: func(t *testing.T) {
				if err := os.MkdirAll(tmpPath("existing"), 0755); err != nil {
					t.Fatalf("Failed to create existing directory: %v", err)
				}
			},
			opts: Options{
				Path: tmpPath("existing"),
			},
			checkResult: func(t *testing.T, got Result) {
				if got.Error != nil {
					t.Errorf("Expected no error for existing directory, got %v", got.Error)
				}
			},
		},
		{
			name: "handle permission denied",
			setup: func() error {
				path := tmpPath("noperm")
				if err := os.MkdirAll(path, 0755); err != nil {
					return err
				}
				return os.Chmod(path, 0000)
			},
			cleanup: func() error {
				return os.Chmod(tmpPath("noperm"), 0755)
			},
			opts: Options{
				Path: tmpPath("noperm/child"),
			},
			wantErr: ErrPermissionDenied,
		},
		{
			name: "handle empty path",
			opts: Options{
				Path: "",
			},
			wantErr: ErrInvalidPath,
		},
		{
			name: "handle git HTTPS URL",
			opts: Options{
				Path: "https://github.com/deblasis/take.git",
			},
			setup: func(t *testing.T) {
				// Skip if no git credentials available
				cmd := exec.Command("git", "config", "--get", "credential.helper")
				if err := cmd.Run(); err != nil {
					t.Skip("Git credentials not configured, skipping clone test")
				}
			},
			checkResult: func(t *testing.T, got Result) {
				if !got.WasCloned {
					t.Error("Expected repository to be cloned")
				}
			},
		},
		{
			name: "handle git SSH URL",
				opts: Options{
					Path: "git@github.com:deblasis/take.git",
				},
				setup: func(t *testing.T) {
					// Skip if no SSH key available
					_, err := os.Stat(filepath.Join(os.Getenv("HOME"), ".ssh", "id_rsa"))
					if err != nil {
						t.Skip("SSH key not found, skipping SSH clone test")
					}
				},
				checkResult: func(t *testing.T, got Result) {
					if !got.WasCloned {
						t.Error("Expected repository to be cloned")
					}
				},
		},
		{
			name: "handle tarball URL",
			opts: Options{
				Path: ts.URL + "/test.tar.gz",
			},
			checkResult: func(t *testing.T, got Result) {
				if !got.WasDownloaded {
					t.Error("Expected tarball to be downloaded")
				}
			},
		},
		{
			name: "handle zip URL",
			opts: Options{
				Path: ts.URL + "/test.zip",
			},
			checkResult: func(t *testing.T, got Result) {
				if !got.WasDownloaded {
					t.Error("Expected zip file to be downloaded")
				}
			},
		},
		{
			name: "handle invalid URL",
			opts: Options{
				Path: "http://invalid.url/file.xyz",
			},
			wantErr: ErrInvalidURL,
		},
		{
			name: "handle download failure",
			opts: Options{
				Path: ts.URL + "/nonexistent.tar.gz",
			},
			wantErr: ErrDownloadFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t)
			}

			got := Take(tt.opts)

			if tt.cleanup != nil {
				if err := tt.cleanup(); err != nil {
					t.Errorf("Cleanup failed: %v", err)
				}
			}

			if tt.wantErr != nil {
				if got.Error != tt.wantErr {
					t.Errorf("Take() error = %v, wantErr %v", got.Error, tt.wantErr)
				}
				return
			}

			if got.Error != nil {
				t.Errorf("Take() unexpected error = %v", got.Error)
				return
			}

			if tt.checkResult != nil {
				tt.checkResult(t, got)
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name: "expand home directory",
			path: "~/test",
			want: filepath.Join(home, "test"),
		},
		{
			name:    "handle empty path",
			path:    "",
			wantErr: true,
		},
		{
			name: "handle absolute path",
			path: "/tmp/test",
			want: "/tmp/test",
		},
		{
			name: "handle relative path",
			path: "./test",
			want: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := expandPath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("expandPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("expandPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
