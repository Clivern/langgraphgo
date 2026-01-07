package tool

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

// MarkdownFile represents a Markdown file with frontmatter and content
type MarkdownFile struct {
	Path        string
	Frontmatter map[string]any
	Content     string
}

// ReadMarkdown reads a Markdown file and parses its frontmatter and content
func ReadMarkdown(filePath string) (*MarkdownFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read markdown file '%s': %w", filePath, err)
	}

	mf := &MarkdownFile{
		Path:        filePath,
		Frontmatter: make(map[string]any),
	}

	// Parse frontmatter (YAML between --- delimiters)
	strContent := string(content)
	if strings.HasPrefix(strContent, "---") {
		endIndex := strings.Index(strContent[3:], "\n---")
		if endIndex > 0 {
			// Extract frontmatter (for now, just store as string)
			// In production, you'd use yaml.Unmarshal here
			frontmatterText := strContent[4 : 4+endIndex]
			mf.Frontmatter["raw"] = frontmatterText
			mf.Content = strings.TrimSpace(strContent[4+endIndex+4:])
			return mf, nil
		}
	}

	// No frontmatter found
	mf.Content = strContent
	return mf, nil
}

// WriteMarkdown writes content to a Markdown file with optional frontmatter
func WriteMarkdown(filePath string, content string, frontmatter map[string]any) error {
	var sb strings.Builder

	// Write frontmatter if present
	if len(frontmatter) > 0 {
		sb.WriteString("---\n")
		for key, value := range frontmatter {
			sb.WriteString(fmt.Sprintf("%s: %v\n", key, value))
		}
		sb.WriteString("---\n\n")
	}

	sb.WriteString(content)

	err := os.WriteFile(filePath, []byte(sb.String()), 0600)
	if err != nil {
		return fmt.Errorf("failed to write markdown file '%s': %w", filePath, err)
	}

	return nil
}

// UpdateMarkdownCheckboxes updates the checkbox status in a Markdown file
func UpdateMarkdownCheckboxes(filePath string, updates map[string]bool) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read markdown file '%s': %w", filePath, err)
	}

	strContent := string(content)
	lines := strings.Split(strContent, "\n")

	// Build pattern for each phase name
	for phaseName, checked := range updates {
		// Look for checkboxes with this phase name
		pattern := regexp.MustCompile(`^- \[ \] Phase \d+:\s+` + regexp.QuoteMeta(phaseName))
		checkedPattern := regexp.MustCompile(`^- \[x\] Phase \d+:\s+` + regexp.QuoteMeta(phaseName))

		for i, line := range lines {
			if checked {
				// Update unchecked to checked
				if pattern.MatchString(line) {
					lines[i] = pattern.ReplaceAllString(line, "- [x] Phase $1: "+phaseName)
					// Need to preserve the phase number
					lines[i] = strings.Replace(line, "- [ ]", "- [x]", 1)
				}
			} else {
				// Update checked to unchecked
				if checkedPattern.MatchString(line) {
					lines[i] = strings.Replace(line, "- [x]", "- [ ]", 1)
				}
			}
		}
	}

	updatedContent := strings.Join(lines, "\n")

	// Write back to file
	err = os.WriteFile(filePath, []byte(updatedContent), 0600)
	if err != nil {
		return "", fmt.Errorf("failed to update markdown file '%s': %w", filePath, err)
	}

	return updatedContent, nil
}

// AppendToMarkdownSection appends content to a specific section in a Markdown file
func AppendToMarkdownSection(filePath, sectionName, content string) error {
	mf, err := ReadMarkdown(filePath)
	if err != nil {
		return err
	}

	// Find the section and append content
	lines := strings.Split(mf.Content, "\n")
	var inSection bool
	var insertIndex int = -1

	for i, line := range lines {
		// Check if this is our section (## SectionName)
		if strings.HasPrefix(line, "## ") && strings.Contains(line, sectionName) {
			inSection = true
			continue
		}

		// If we're in a section and hit another section header
		if inSection && strings.HasPrefix(line, "#") {
			insertIndex = i
			break
		}
	}

	// If section not found, create it
	if insertIndex == -1 {
		if inSection {
			// Section exists but is at the end of file
			lines = append(lines, "", content)
		} else {
			// Section doesn't exist, create it
			lines = append(lines, "", fmt.Sprintf("## %s", sectionName), "", content)
		}
	} else {
		// Insert at the found position
		newLines := make([]string, 0, len(lines)+1)
		newLines = append(newLines, lines[:insertIndex]...)
		newLines = append(newLines, content)
		newLines = append(newLines, lines[insertIndex:]...)
		lines = newLines
	}

	// Write back
	mf.Content = strings.Join(lines, "\n")
	return WriteMarkdown(filePath, mf.Content, mf.Frontmatter)
}

// LogErrorToMarkdown logs an error with timestamp to a Markdown file
func LogErrorToMarkdown(filePath, errorMessage string) error {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Create error entry
	errorEntry := fmt.Sprintf("\n## Error [%s]\n%s\n", timestamp, errorMessage)

	// Append to file
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open markdown file '%s': %w", filePath, err)
	}
	defer f.Close()

	_, err = f.WriteString(errorEntry)
	if err != nil {
		return fmt.Errorf("failed to write error to markdown file '%s': %w", filePath, err)
	}

	return nil
}

// ParseTaskPlan extracts phases and goals from a task plan Markdown file
func ParseTaskPlan(filePath string) (goal string, phases []TaskPhase, err error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", nil, fmt.Errorf("failed to read task plan '%s': %w", filePath, err)
	}

	lines := strings.Split(string(content), "\n")
	var inGoalSection bool
	var inPhasesSection bool
	var goalBuilder strings.Builder

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for goal section
		if strings.HasPrefix(line, "%% Goal") {
			inGoalSection = true
			inPhasesSection = false
			continue
		}

		// Check for phases section
		if strings.HasPrefix(line, "%% Phases") {
			inPhasesSection = true
			inGoalSection = false
			continue
		}

		// Parse goal
		if inGoalSection && line != "" {
			if goalBuilder.Len() > 0 {
				goalBuilder.WriteString(" ")
			}
			goalBuilder.WriteString(line)
		}

		// Parse phases
		if inPhasesSection {
			if strings.HasPrefix(line, "- [") {
				phase := parsePhaseLine(line)
				if phase.Name != "" {
					phases = append(phases, phase)
				}
			}
		}
	}

	return goalBuilder.String(), phases, nil
}

// TaskPhase represents a single phase in a task plan
type TaskPhase struct {
	Number      int
	Name        string
	Description string
	Node        string
	Complete    bool
}

func parsePhaseLine(line string) TaskPhase {
	phase := TaskPhase{}

	// Extract checkbox status
	phase.Complete = strings.HasPrefix(line, "- [x]")

	// Extract phase number and name
	// Format: - [x] Phase 1: Research
	re := regexp.MustCompile(`Phase (\d+):\s+(.+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		_, _ = fmt.Sscanf(matches[1], "%d", &phase.Number)
		phase.Name = matches[2]
	}

	return phase
}

// GenerateTaskPlanMarkdown creates a task plan Markdown content from goal and phases
func GenerateTaskPlanMarkdown(goal string, phases []TaskPhase) string {
	var sb strings.Builder

	sb.WriteString("%% Goal\n\n")
	sb.WriteString(goal)
	sb.WriteString("\n\n%% Phases\n\n")

	for _, phase := range phases {
		if phase.Complete {
			sb.WriteString(fmt.Sprintf("- [x] Phase %d: %s\n", phase.Number, phase.Name))
		} else {
			sb.WriteString(fmt.Sprintf("- [ ] Phase %d: %s\n", phase.Number, phase.Name))
		}
		if phase.Description != "" {
			sb.WriteString(fmt.Sprintf("  Description: %s\n", phase.Description))
		}
		if phase.Node != "" {
			sb.WriteString(fmt.Sprintf("  Node: %s\n", phase.Node))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// ExtractSectionContent extracts content from a specific section in a Markdown file
func ExtractSectionContent(filePath, sectionName string) (string, error) {
	mf, err := ReadMarkdown(filePath)
	if err != nil {
		return "", err
	}

	lines := strings.Split(mf.Content, "\n")
	var content strings.Builder
	var inSection bool

	for _, line := range lines {
		// Check if this is our section
		if strings.HasPrefix(line, "## ") && strings.Contains(line, sectionName) {
			inSection = true
			continue
		}

		// If we're in a section and hit another section header, stop
		if inSection && strings.HasPrefix(line, "#") {
			break
		}

		// Collect content
		if inSection && line != "" {
			if content.Len() > 0 {
				content.WriteString("\n")
			}
			content.WriteString(line)
		}
	}

	return content.String(), nil
}

// CreateTaskPlan creates a new task plan file with goal and phases
func CreateTaskPlan(filePath, goal string, phases []TaskPhase) error {
	content := GenerateTaskPlanMarkdown(goal, phases)
	return WriteFile(filePath, content)
}

// UpdatePhaseStatus updates the status (complete/incomplete) of a specific phase
func UpdatePhaseStatus(filePath, phaseName string, complete bool) error {
	updates := map[string]bool{
		phaseName: complete,
	}

	_, err := UpdateMarkdownCheckboxes(filePath, updates)
	return err
}

// GetCompletedPhases returns a list of completed phase names from a task plan
func GetCompletedPhases(filePath string) ([]string, error) {
	_, phases, err := ParseTaskPlan(filePath)
	if err != nil {
		return nil, err
	}

	var completed []string
	for _, phase := range phases {
		if phase.Complete {
			completed = append(completed, phase.Name)
		}
	}

	return completed, nil
}

// GetPendingPhases returns a list of pending (incomplete) phase names from a task plan
func GetPendingPhases(filePath string) ([]string, error) {
	_, phases, err := ParseTaskPlan(filePath)
	if err != nil {
		return nil, err
	}

	var pending []string
	for _, phase := range phases {
		if !phase.Complete {
			pending = append(pending, phase.Name)
		}
	}

	return pending, nil
}
