package tool

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadMarkdown(t *testing.T) {
	t.Run("read simple markdown", func(t *testing.T) {
		// Create temporary file
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.md")

		content := "# Test Document\n\nThis is a test."
		err := os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Read markdown
		mf, err := ReadMarkdown(filePath)
		if err != nil {
			t.Fatalf("ReadMarkdown() error = %v", err)
		}

		if mf.Content != content {
			t.Errorf("Content mismatch, got %q", mf.Content)
		}
	})

	t.Run("read markdown with frontmatter", func(t *testing.T) {
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "test.md")

		content := "---\ntitle: Test\ndate: 2025-01-07\n---\n\n# Content"
		err := os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		mf, err := ReadMarkdown(filePath)
		if err != nil {
			t.Fatalf("ReadMarkdown() error = %v", err)
		}

		if mf.Frontmatter == nil {
			t.Error("Frontmatter not parsed")
		}

		if mf.Content != "# Content" {
			t.Errorf("Content not extracted correctly, got %q", mf.Content)
		}
	})
}

func TestWriteMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.md")

	content := "# Test Document\n\nThis is test content."
	frontmatter := map[string]any{
		"title": "Test",
		"date":  "2025-01-07",
	}

	err := WriteMarkdown(filePath, content, frontmatter)
	if err != nil {
		t.Fatalf("WriteMarkdown() error = %v", err)
	}

	// Read back and verify
	mf, err := ReadMarkdown(filePath)
	if err != nil {
		t.Fatalf("Failed to read back: %v", err)
	}

	if !strings.Contains(mf.Content, "Test Document") {
		t.Error("Content not written correctly")
	}
}

func TestUpdateMarkdownCheckboxes(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plan.md")

	// Create initial plan
	initialContent := `%% Goal
Test goal

%% Phases
- [ ] Phase 1: Research
- [ ] Phase 2: Write
- [ ] Phase 3: Review
`
	err := os.WriteFile(filePath, []byte(initialContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Update checkboxes
	updates := map[string]bool{
		"Research": true,
		"Write":    true,
	}

	_, err = UpdateMarkdownCheckboxes(filePath, updates)
	if err != nil {
		t.Fatalf("UpdateMarkdownCheckboxes() error = %v", err)
	}

	// Verify updates
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "- [x] Phase 1: Research") {
		t.Error("Phase 1 not marked as complete")
	}

	if !strings.Contains(strContent, "- [x] Phase 2: Write") {
		t.Error("Phase 2 not marked as complete")
	}

	if !strings.Contains(strContent, "- [ ] Phase 3: Review") {
		t.Error("Phase 3 should remain incomplete")
	}
}

func TestAppendToMarkdownSection(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notes.md")

	// Create file with sections
	initialContent := `# Notes

## Research
Initial notes

## Errors
`
	err := os.WriteFile(filePath, []byte(initialContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Append to Research section
	newContent := "Additional research findings"
	err = AppendToMarkdownSection(filePath, "Research", newContent)
	if err != nil {
		t.Fatalf("AppendToMarkdownSection() error = %v", err)
	}

	// Verify
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read back: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "Additional research findings") {
		t.Error("Content not appended")
	}
}

func TestLogErrorToMarkdown(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notes.md")

	// Log error
	err := LogErrorToMarkdown(filePath, "Test error message")
	if err != nil {
		t.Fatalf("LogErrorToMarkdown() error = %v", err)
	}

	// Verify
	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed to read back: %v", err)
	}

	strContent := string(content)
	if !strings.Contains(strContent, "## Error [") {
		t.Error("Error header not found")
	}

	if !strings.Contains(strContent, "Test error message") {
		t.Error("Error message not found")
	}
}

func TestParseTaskPlan(t *testing.T) {
	t.Run("parse complete plan", func(t *testing.T) {
		content := `%% Goal
Research TypeScript benefits

%% Phases
- [ ] Phase 1: Research
  Description: Search for information
  Node: research

- [x] Phase 2: Compile
  Description: Compile findings
  Node: compile
`
		tmpDir := t.TempDir()
		filePath := filepath.Join(tmpDir, "plan.md")
		err := os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		goal, phases, err := ParseTaskPlan(filePath)
		if err != nil {
			t.Fatalf("ParseTaskPlan() error = %v", err)
		}

		if goal != "Research TypeScript benefits" {
			t.Errorf("Goal = %q, want 'Research TypeScript benefits'", goal)
		}

		if len(phases) != 2 {
			t.Fatalf("Got %d phases, want 2", len(phases))
		}

		if phases[0].Complete {
			t.Error("Phase 1 should be incomplete")
		}

		if !phases[1].Complete {
			t.Error("Phase 2 should be complete")
		}
	})
}

func TestGenerateTaskPlanMarkdown(t *testing.T) {
	phases := []TaskPhase{
		{
			Number:      1,
			Name:        "Research",
			Description: "Search for information",
			Node:        "research",
			Complete:    false,
		},
		{
			Number:      2,
			Name:        "Compile",
			Description: "Compile findings",
			Node:        "compile",
			Complete:    true,
		},
	}

	markdown := GenerateTaskPlanMarkdown("Test goal", phases)

	if !strings.Contains(markdown, "%% Goal") {
		t.Error("Missing goal section")
	}

	if !strings.Contains(markdown, "%% Phases") {
		t.Error("Missing phases section")
	}

	if !strings.Contains(markdown, "- [ ] Phase 1: Research") {
		t.Error("Phase 1 not marked as incomplete")
	}

	if !strings.Contains(markdown, "- [x] Phase 2: Compile") {
		t.Error("Phase 2 not marked as complete")
	}

	if !strings.Contains(markdown, "Description: Search for information") {
		t.Error("Phase 1 description missing")
	}

	if !strings.Contains(markdown, "Node: research") {
		t.Error("Phase 1 node missing")
	}
}

func TestCreateTaskPlan(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plan.md")

	phases := []TaskPhase{
		{
			Number:      1,
			Name:        "Research",
			Description: "Research phase",
			Node:        "research",
		},
	}

	err := CreateTaskPlan(filePath, "Test goal", phases)
	if err != nil {
		t.Fatalf("CreateTaskPlan() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("Plan file not created")
	}
}

func TestUpdatePhaseStatus(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plan.md")

	// Create initial plan
	phases := []TaskPhase{
		{Number: 1, Name: "Research"},
		{Number: 2, Name: "Write"},
	}
	err := CreateTaskPlan(filePath, "Test goal", phases)
	if err != nil {
		t.Fatalf("Failed to create plan: %v", err)
	}

	// Update phase status
	err = UpdatePhaseStatus(filePath, "Research", true)
	if err != nil {
		t.Fatalf("UpdatePhaseStatus() error = %v", err)
	}

	// Verify
	completed, err := GetCompletedPhases(filePath)
	if err != nil {
		t.Fatalf("GetCompletedPhases() error = %v", err)
	}

	if len(completed) != 1 || completed[0] != "Research" {
		t.Errorf("Got completed phases %v, want [Research]", completed)
	}
}

func TestGetCompletedPhases(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plan.md")

	content := `%% Goal
Test

%% Phases
- [x] Phase 1: Research
- [ ] Phase 2: Write
`
	err := os.WriteFile(filePath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	completed, err := GetCompletedPhases(filePath)
	if err != nil {
		t.Fatalf("GetCompletedPhases() error = %v", err)
	}

	if len(completed) != 1 {
		t.Errorf("Got %d completed phases, want 1", len(completed))
	}

	if completed[0] != "Research" {
		t.Errorf("Got phase %q, want 'Research'", completed[0])
	}
}

func TestGetPendingPhases(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "plan.md")

	content := `%% Goal
Test

%% Phases
- [x] Phase 1: Research
- [ ] Phase 2: Write
- [ ] Phase 3: Review
`
	err := os.WriteFile(filePath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	pending, err := GetPendingPhases(filePath)
	if err != nil {
		t.Fatalf("GetPendingPhases() error = %v", err)
	}

	if len(pending) != 2 {
		t.Errorf("Got %d pending phases, want 2", len(pending))
	}
}

func TestExtractSectionContent(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "notes.md")

	content := `# Notes

## Research
Research finding 1
Research finding 2

## Errors
Error log
`
	err := os.WriteFile(filePath, []byte(content), 0600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	extracted, err := ExtractSectionContent(filePath, "Research")
	if err != nil {
		t.Fatalf("ExtractSectionContent() error = %v", err)
	}

	if !strings.Contains(extracted, "Research finding 1") {
		t.Error("Expected content not extracted")
	}

	if strings.Contains(extracted, "Error log") {
		t.Error("Extracted content from wrong section")
	}
}

func TestParsePhaseLine(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected TaskPhase
	}{
		{
			name: "incomplete phase",
			line: "- [ ] Phase 1: Research",
			expected: TaskPhase{
				Number:   1,
				Name:     "Research",
				Complete: false,
			},
		},
		{
			name: "complete phase",
			line: "- [x] Phase 2: Write Summary",
			expected: TaskPhase{
				Number:   2,
				Name:     "Write Summary",
				Complete: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePhaseLine(tt.line)
			if result.Number != tt.expected.Number {
				t.Errorf("Number = %d, want %d", result.Number, tt.expected.Number)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %q, want %q", result.Name, tt.expected.Name)
			}
			if result.Complete != tt.expected.Complete {
				t.Errorf("Complete = %v, want %v", result.Complete, tt.expected.Complete)
			}
		})
	}
}
