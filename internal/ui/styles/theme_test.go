package styles

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()

	// Test that all theme components produce output (indicating they're initialized)
	// We can't compare lipgloss.Style directly, so we test by rendering
	testText := "test"

	docResult := theme.Doc.Render(testText)
	if docResult == "" {
		t.Error("Expected Doc style to produce output")
	}

	titleResult := theme.Title.Render(testText)
	if titleResult == "" {
		t.Error("Expected Title style to produce output")
	}

	helpResult := theme.Help.Render(testText)
	if helpResult == "" {
		t.Error("Expected Help style to produce output")
	}

	searchResult := theme.Search.Render(testText)
	if searchResult == "" {
		t.Error("Expected Search style to produce output")
	}

	// Test specific style properties
	t.Run("Doc style", func(t *testing.T) {
		// Test that margin is set
		rendered := theme.Doc.Render("test")
		if rendered == "test" {
			t.Error("Expected Doc style to apply margin formatting")
		}
	})

	t.Run("Title style", func(t *testing.T) {
		// Test that title has color and bold formatting
		rendered := theme.Title.Render("Test Title")
		if rendered == "Test Title" {
			t.Error("Expected Title style to apply formatting")
		}
	})

	t.Run("Help style", func(t *testing.T) {
		// Test that help style is applied
		rendered := theme.Help.Render("help text")
		if rendered == "help text" {
			t.Error("Expected Help style to apply formatting")
		}
	})

	t.Run("Search style", func(t *testing.T) {
		// Test that search style includes border
		rendered := theme.Search.Render("search content")
		if rendered == "search content" {
			t.Error("Expected Search style to apply border and formatting")
		}
	})
}

func TestDefaultThemeConsistency(t *testing.T) {
	// Test that calling DefaultTheme() multiple times returns consistent themes
	theme1 := DefaultTheme()
	theme2 := DefaultTheme()

	// While we can't directly compare lipgloss.Style objects,
	// we can test that they produce the same rendered output
	testText := "test"

	if theme1.Doc.Render(testText) != theme2.Doc.Render(testText) {
		t.Error("DefaultTheme() should return consistent Doc styles")
	}

	if theme1.Title.Render(testText) != theme2.Title.Render(testText) {
		t.Error("DefaultTheme() should return consistent Title styles")
	}

	if theme1.Help.Render(testText) != theme2.Help.Render(testText) {
		t.Error("DefaultTheme() should return consistent Help styles")
	}

	if theme1.Search.Render(testText) != theme2.Search.Render(testText) {
		t.Error("DefaultTheme() should return consistent Search styles")
	}
}

func TestDefaultTableTheme(t *testing.T) {
	tableTheme := DefaultTableTheme()

	// Test that all table theme colors are set
	if tableTheme.HeaderBorderColor == "" {
		t.Error("Expected HeaderBorderColor to be set")
	}

	if tableTheme.SelectedFg == "" {
		t.Error("Expected SelectedFg to be set")
	}

	if tableTheme.SelectedBg == "" {
		t.Error("Expected SelectedBg to be set")
	}

	// Test specific color values
	expectedHeaderBorderColor := lipgloss.Color("240")
	if tableTheme.HeaderBorderColor != expectedHeaderBorderColor {
		t.Errorf("Expected HeaderBorderColor to be %v, got %v", expectedHeaderBorderColor, tableTheme.HeaderBorderColor)
	}

	expectedSelectedFg := lipgloss.Color("229")
	if tableTheme.SelectedFg != expectedSelectedFg {
		t.Errorf("Expected SelectedFg to be %v, got %v", expectedSelectedFg, tableTheme.SelectedFg)
	}

	expectedSelectedBg := lipgloss.Color("57")
	if tableTheme.SelectedBg != expectedSelectedBg {
		t.Errorf("Expected SelectedBg to be %v, got %v", expectedSelectedBg, tableTheme.SelectedBg)
	}
}

func TestDefaultTableThemeConsistency(t *testing.T) {
	// Test that calling DefaultTableTheme() multiple times returns consistent themes
	theme1 := DefaultTableTheme()
	theme2 := DefaultTableTheme()

	if theme1.HeaderBorderColor != theme2.HeaderBorderColor {
		t.Error("DefaultTableTheme() should return consistent HeaderBorderColor")
	}

	if theme1.SelectedFg != theme2.SelectedFg {
		t.Error("DefaultTableTheme() should return consistent SelectedFg")
	}

	if theme1.SelectedBg != theme2.SelectedBg {
		t.Error("DefaultTableTheme() should return consistent SelectedBg")
	}
}

func TestThemeStructCreation(t *testing.T) {
	// Test manual creation of Theme struct
	doc := lipgloss.NewStyle().Margin(2, 3)
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("200"))
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))
	search := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	theme := Theme{
		Doc:    doc,
		Title:  title,
		Help:   help,
		Search: search,
	}

	// Test that all fields are properly set by checking rendered output
	testContent := "test content"
	docResult := theme.Doc.Render(testContent)
	titleResult := theme.Title.Render(testContent)
	helpResult := theme.Help.Render(testContent)
	searchResult := theme.Search.Render(testContent)

	// Test that we get the expected rendered results from our custom styles
	expectedDocResult := doc.Render(testContent)
	expectedTitleResult := title.Render(testContent)
	expectedHelpResult := help.Render(testContent)
	expectedSearchResult := search.Render(testContent)

	if docResult != expectedDocResult {
		t.Error("Doc style not properly assigned")
	}

	if titleResult != expectedTitleResult {
		t.Error("Title style not properly assigned")
	}

	if helpResult != expectedHelpResult {
		t.Error("Help style not properly assigned")
	}

	if searchResult != expectedSearchResult {
		t.Error("Search style not properly assigned")
	}

	// Test that styles work correctly
	if theme.Doc.Render(testContent) == testContent {
		t.Error("Expected Doc style to apply formatting")
	}
}

func TestTableThemeStructCreation(t *testing.T) {
	// Test manual creation of TableTheme struct
	headerColor := lipgloss.Color("100")
	selectedFg := lipgloss.Color("200")
	selectedBg := lipgloss.Color("50")

	tableTheme := TableTheme{
		HeaderBorderColor: headerColor,
		SelectedFg:        selectedFg,
		SelectedBg:        selectedBg,
	}

	// Test that all fields are properly set
	if tableTheme.HeaderBorderColor != headerColor {
		t.Error("HeaderBorderColor not properly assigned")
	}

	if tableTheme.SelectedFg != selectedFg {
		t.Error("SelectedFg not properly assigned")
	}

	if tableTheme.SelectedBg != selectedBg {
		t.Error("SelectedBg not properly assigned")
	}
}

func TestThemeColorsAreValid(t *testing.T) {
	// Test that default theme colors are valid lipgloss colors
	tableTheme := DefaultTableTheme()

	// These should not panic when used with lipgloss
	testStyle := lipgloss.NewStyle().
		BorderForeground(tableTheme.HeaderBorderColor).
		Foreground(tableTheme.SelectedFg).
		Background(tableTheme.SelectedBg)

	// Test that the style can render without errors
	result := testStyle.Render("test")
	if result == "" {
		t.Error("Expected non-empty rendered result with theme colors")
	}
}

func TestThemeStylesCanBeComposed(t *testing.T) {
	theme := DefaultTheme()

	// Test that styles can be composed and modified
	composedStyle := theme.Title.
		Margin(1).
		Border(lipgloss.RoundedBorder())

	testText := "composed test"
	result := composedStyle.Render(testText)

	if result == testText {
		t.Error("Expected composed style to apply formatting")
	}

	// Original style should remain unchanged
	originalResult := theme.Title.Render(testText)
	if result == originalResult {
		t.Error("Composed style should be different from original")
	}
}

func TestThemeWithEmptyContent(t *testing.T) {
	theme := DefaultTheme()

	// Test that theme styles handle empty content gracefully
	emptyResult := theme.Doc.Render("")
	if emptyResult == "" {
		// This might be expected behavior, but we want to ensure it doesn't panic
		t.Log("Doc style with empty content produces empty result")
	}

	// Test with whitespace
	whitespaceResult := theme.Title.Render("   ")
	if len(whitespaceResult) < 3 {
		t.Error("Expected Title style to preserve whitespace content")
	}
}

func TestThemeWithSpecialCharacters(t *testing.T) {
	theme := DefaultTheme()

	testCases := []struct {
		name    string
		content string
	}{
		{"Unicode emojis", "🎉 📋 🚀"},
		{"Special chars", "!@#$%^&*()"},
		{"Newlines", "line1\nline2\nline3"},
		{"Tabs", "col1\tcol2\tcol3"},
		{"Mixed", "Hello 🌍\nTab:\there\t!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that all theme styles can handle special characters
			docResult := theme.Doc.Render(tc.content)
			titleResult := theme.Title.Render(tc.content)
			helpResult := theme.Help.Render(tc.content)
			searchResult := theme.Search.Render(tc.content)

			// None should panic or return empty (unless the style specifically handles it that way)
			if docResult == "" && tc.content != "" {
				t.Errorf("Doc style returned empty result for content: %q", tc.content)
			}

			if titleResult == "" && tc.content != "" {
				t.Errorf("Title style returned empty result for content: %q", tc.content)
			}

			if helpResult == "" && tc.content != "" {
				t.Errorf("Help style returned empty result for content: %q", tc.content)
			}

			if searchResult == "" && tc.content != "" {
				t.Errorf("Search style returned empty result for content: %q", tc.content)
			}
		})
	}
}

func TestTableThemeColorValues(t *testing.T) {
	// Test that table theme uses appropriate color values
	tableTheme := DefaultTableTheme()

	// Test color string representations
	headerColorStr := string(tableTheme.HeaderBorderColor)
	if headerColorStr != "240" {
		t.Errorf("Expected HeaderBorderColor to be '240', got %q", headerColorStr)
	}

	selectedFgStr := string(tableTheme.SelectedFg)
	if selectedFgStr != "229" {
		t.Errorf("Expected SelectedFg to be '229', got %q", selectedFgStr)
	}

	selectedBgStr := string(tableTheme.SelectedBg)
	if selectedBgStr != "57" {
		t.Errorf("Expected SelectedBg to be '57', got %q", selectedBgStr)
	}
}

func TestZeroValueThemes(t *testing.T) {
	// Test behavior with zero-value themes
	var theme Theme
	var tableTheme TableTheme

	// Zero-value theme should not panic when used
	result := theme.Doc.Render("test")
	if result != "test" {
		// Zero-value lipgloss.Style should render content as-is
		t.Errorf("Expected zero-value Doc style to render 'test', got %q", result)
	}

	// Zero-value table theme should have empty colors
	if tableTheme.HeaderBorderColor != "" {
		t.Error("Zero-value TableTheme should have empty HeaderBorderColor")
	}

	if tableTheme.SelectedFg != "" {
		t.Error("Zero-value TableTheme should have empty SelectedFg")
	}

	if tableTheme.SelectedBg != "" {
		t.Error("Zero-value TableTheme should have empty SelectedBg")
	}
}
