package take

import (
	"os"
	"path/filepath"
	"testing"
)

type testCase struct {
	name        string
	opts        Options
	setup       func() error
	cleanup     func() error
	wantResult  Result
	wantErr     error
	checkResult func(t *testing.T, got Result)
}

func TestTake(t *testing.T) {
	// Create a temporary test directory
	tmpDir, err := os.MkdirTemp("", "take-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
			setup: func() error {
				return os.MkdirAll(tmpPath("existing"), 0755)
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
			name: "detect HTTPS git URL",
			opts: Options{
				Path: "https://github.com/user/repo.git",
			},
			checkResult: func(t *testing.T, got Result) {
				if got.Error == nil || got.Error.Error() != "git clone not implemented" {
					t.Error("Expected git clone not implemented error")
				}
			},
		},
		{
			name: "detect SSH git URL",
			opts: Options{
				Path: "git@github.com:user/repo.git",
			},
			checkResult: func(t *testing.T, got Result) {
				if got.Error == nil || got.Error.Error() != "git clone not implemented" {
					t.Error("Expected git clone not implemented error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				if err := tt.setup(); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}

			if tt.cleanup != nil {
				defer tt.cleanup()
			}

			got := Take(tt.opts)

			if tt.wantErr != nil {
				if got.Error != tt.wantErr {
					t.Errorf("Take() error = %v, wantErr %v", got.Error, tt.wantErr)
				}
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
