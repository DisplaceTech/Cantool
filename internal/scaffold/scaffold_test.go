package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScaffoldBasicTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	err := Scaffold("my-project", "basic", tmpDir)
	if err != nil {
		t.Fatalf("Scaffold returned error: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "my-project")

	// Verify expected files exist.
	expectedFiles := []string{
		"cantool.yaml",
		"daml.yaml",
		"daml/Main.daml",
		"test/Test.daml",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(projectDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", f)
		}
	}

	// Verify no .tmpl files remain.
	err = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, ".tmpl") {
			t.Errorf("found .tmpl file that should have been stripped: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking project dir: %v", err)
	}

	// Verify template rendering in cantool.yaml.
	content, err := os.ReadFile(filepath.Join(projectDir, "cantool.yaml"))
	if err != nil {
		t.Fatalf("reading cantool.yaml: %v", err)
	}
	if !strings.Contains(string(content), "name: my-project") {
		t.Error("cantool.yaml does not contain project name")
	}
	if !strings.Contains(string(content), "sdk-version: "+DefaultSDKVersion) {
		t.Error("cantool.yaml does not contain SDK version")
	}

	// Verify template rendering in daml.yaml.
	damlYaml, err := os.ReadFile(filepath.Join(projectDir, "daml.yaml"))
	if err != nil {
		t.Fatalf("reading daml.yaml: %v", err)
	}
	if !strings.Contains(string(damlYaml), "name: my-project") {
		t.Error("daml.yaml does not contain project name")
	}
	if !strings.Contains(string(damlYaml), "version: 0.0.1") {
		t.Error("daml.yaml does not contain version")
	}
}

func TestScaffoldTokenTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	err := Scaffold("token-app", "token", tmpDir)
	if err != nil {
		t.Fatalf("Scaffold returned error: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "token-app")

	expectedFiles := []string{
		"cantool.yaml",
		"daml.yaml",
		"daml/Token.daml",
		"test/TokenTest.daml",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(projectDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s does not exist", f)
		}
	}

	// Verify token cantool.yaml has parties.
	content, err := os.ReadFile(filepath.Join(projectDir, "cantool.yaml"))
	if err != nil {
		t.Fatalf("reading cantool.yaml: %v", err)
	}
	if !strings.Contains(string(content), "Alice") {
		t.Error("token cantool.yaml does not contain Alice party")
	}
	if !strings.Contains(string(content), "Bob") {
		t.Error("token cantool.yaml does not contain Bob party")
	}

	// Verify Token.daml has Transfer, Mint, Burn choices.
	tokenDaml, err := os.ReadFile(filepath.Join(projectDir, "daml/Token.daml"))
	if err != nil {
		t.Fatalf("reading Token.daml: %v", err)
	}
	for _, choice := range []string{"Transfer", "Mint", "Burn"} {
		if !strings.Contains(string(tokenDaml), "choice "+choice) {
			t.Errorf("Token.daml does not contain choice %s", choice)
		}
	}
}

func TestScaffoldInvalidName(t *testing.T) {
	tests := []struct {
		name string
	}{
		{""},
		{"has spaces"},
		{"-starts-with-hyphen"},
		{"has@special"},
		{"has.dots"},
		{"has/slashes"},
	}

	for _, tt := range tests {
		t.Run("name="+tt.name, func(t *testing.T) {
			err := Scaffold(tt.name, "basic", t.TempDir())
			if err == nil {
				t.Errorf("expected error for invalid name %q, got nil", tt.name)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid project name") {
				t.Errorf("expected 'invalid project name' error, got: %v", err)
			}
		})
	}
}

func TestScaffoldValidNames(t *testing.T) {
	tests := []string{
		"myproject",
		"my-project",
		"My-Project",
		"project123",
		"123project",
		"a",
		"A",
		"1",
	}

	for _, name := range tests {
		t.Run("name="+name, func(t *testing.T) {
			err := Scaffold(name, "basic", t.TempDir())
			if err != nil {
				t.Errorf("expected no error for valid name %q, got: %v", name, err)
			}
		})
	}
}

func TestScaffoldDirectoryAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "existing")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		t.Fatal(err)
	}

	err := Scaffold("existing", "basic", tmpDir)
	if err == nil {
		t.Fatal("expected error when directory exists, got nil")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' error, got: %v", err)
	}
}

func TestScaffoldUnknownTemplate(t *testing.T) {
	err := Scaffold("test-project", "nonexistent", t.TempDir())
	if err == nil {
		t.Fatal("expected error for unknown template, got nil")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("expected 'unknown template' error, got: %v", err)
	}
}

func TestScaffoldTmplExtensionStripped(t *testing.T) {
	tmpDir := t.TempDir()
	if err := Scaffold("strip-test", "basic", tmpDir); err != nil {
		t.Fatalf("Scaffold returned error: %v", err)
	}

	projectDir := filepath.Join(tmpDir, "strip-test")
	var tmplFound bool
	filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.HasSuffix(info.Name(), ".tmpl") {
			tmplFound = true
		}
		return nil
	})
	if tmplFound {
		t.Error("found .tmpl extension in output files; extensions should be stripped")
	}
}
