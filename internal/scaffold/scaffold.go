package scaffold

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// DefaultSDKVersion is the default Canton SDK version used in scaffolded projects.
const DefaultSDKVersion = "3.4.11"

// templateData holds the values available in template rendering.
type templateData struct {
	Name       string
	SDKVersion string
}

// validName matches project names: must start with a letter or digit,
// then letters, digits, or hyphens. No spaces.
var validName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*$`)

// Scaffold creates a new project directory at targetDir/<name> using the
// named template set. It validates the project name, checks the target
// does not already exist, renders all templates, and writes the output
// files with the .tmpl extension stripped.
func Scaffold(name, templateName, targetDir string) error {
	if !validName.MatchString(name) {
		return fmt.Errorf("invalid project name %q: must start with a letter or digit and contain only letters, digits, or hyphens", name)
	}

	projectDir := filepath.Join(targetDir, name)
	if _, err := os.Stat(projectDir); err == nil {
		return fmt.Errorf("directory %q already exists", projectDir)
	}

	templateRoot := filepath.Join("templates", templateName)

	// Verify the template exists in the embedded FS.
	if _, err := fs.Stat(templateFS, templateRoot); err != nil {
		return fmt.Errorf("unknown template %q", templateName)
	}

	data := templateData{
		Name:       name,
		SDKVersion: DefaultSDKVersion,
	}

	return fs.WalkDir(templateFS, templateRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Compute the relative path from the template root.
		relPath, err := filepath.Rel(templateRoot, path)
		if err != nil {
			return err
		}

		// Target path inside the project directory.
		outPath := filepath.Join(projectDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0o755)
		}

		// Read the template file from the embedded FS.
		content, err := fs.ReadFile(templateFS, path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", path, err)
		}

		// Parse and execute the template.
		tmpl, err := template.New(filepath.Base(path)).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", path, err)
		}

		var buf strings.Builder
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("executing template %s: %w", path, err)
		}

		// Strip .tmpl extension from the output file name.
		outPath = strings.TrimSuffix(outPath, ".tmpl")

		// Ensure parent directory exists.
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}

		return os.WriteFile(outPath, []byte(buf.String()), 0o644)
	})
}
