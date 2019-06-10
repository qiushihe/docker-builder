package builder

import "testing"

func TestOptions(t *testing.T) {
	t.Run("options parsing", func(t *testing.T) {
		t.Run("parse source directory option", func(t *testing.T) {
			opts, err := parseOptions([]string{"", "-src", "path/to/src"})
			if err != nil {
				t.Fatalf("failed to parse options: %+v", err)
			}

			if opts.srcPath != "path/to/src" {
				t.Fatalf("unexpected source directory option: %+v", opts.srcPath)
			}
		})

		t.Run("parse lib directory option", func(t *testing.T) {
			opts, err := parseOptions([]string{"", "-lib", "path/to/lib"})
			if err != nil {
				t.Fatalf("failed to parse options: %+v", err)
			}

			if len(opts.libPaths) != 1 {
				t.Fatalf("unexpected number of lib directory options: %+v", len(opts.libPaths))
			}

			if opts.libPaths[0] != "path/to/lib" {
				t.Fatalf("unexpected lib directory option: %+v", opts.libPaths[0])
			}
		})

		t.Run("parse multiple lib directory options", func(t *testing.T) {
			opts, err := parseOptions([]string{"", "-lib", "path/to/lib1", "-lib", "path/to/lib2"})
			if err != nil {
				t.Fatalf("failed to parse options: %+v", err)
			}

			if len(opts.libPaths) != 2 {
				t.Fatalf("unexpected number of lib directory options: %+v", len(opts.libPaths))
			}

			if opts.libPaths[0] != "path/to/lib1" {
				t.Fatalf("unexpected first lib directory option: %+v", opts.libPaths[0])
			}

			if opts.libPaths[1] != "path/to/lib2" {
				t.Fatalf("unexpected second lib directory option: %+v", opts.libPaths[1])
			}
		})
	})

	t.Run("options validation", func(t *testing.T) {
		t.Run("validate presence of source directory option", func(t *testing.T) {
			err := validateOptions(&builderOptions{buildOpts: "build-opts"})
			if err == nil {
				t.Fatalf("should have invalidated lack of source directory option")
			}
		})

		t.Run("validate presence of build options", func(t *testing.T) {
			err := validateOptions(&builderOptions{srcPath: "src-path"})
			if err == nil {
				t.Fatalf("should have invalidated lack of build options")
			}
		})
	})
}
