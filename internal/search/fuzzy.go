package search

import (
	"strings"

	"github.com/bvdwalt/clippy/internal/history"
)

// FuzzyMatcher provides fuzzy search functionality similar to fzf
type FuzzyMatcher struct{}

// NewFuzzyMatcher creates a new fuzzy matcher
func NewFuzzyMatcher() *FuzzyMatcher {
	return &FuzzyMatcher{}
}

// ScoredItem represents an item with its fuzzy match score
type ScoredItem struct {
	Item  history.ClipboardHistory
	Score int
}

// Search performs fuzzy search on clipboard history items
func (f *FuzzyMatcher) Search(items []history.ClipboardHistory, query string) []history.ClipboardHistory {
	if query == "" {
		return nil
	}

	query = strings.ToLower(query)

	var matches []ScoredItem

	for _, item := range items {
		score := f.fuzzyMatch(strings.ToLower(item.Item), query)
		if score > 0 {
			matches = append(matches, ScoredItem{Item: item, Score: score})
		}
	}

	f.sortByScore(matches)

	result := make([]history.ClipboardHistory, len(matches))
	for i, match := range matches {
		result[i] = match.Item
	}

	return result
}

// fuzzyMatch implements fuzzy matching similar to fzf
// Returns a score > 0 if the query matches, 0 if no match
func (f *FuzzyMatcher) fuzzyMatch(text, query string) int {
	if len(query) == 0 {
		return 1
	}
	if len(text) == 0 {
		return 0
	}

	score := 0
	textIdx := 0
	queryIdx := 0
	lastMatchIdx := -1
	consecutiveMatches := 0

	for queryIdx < len(query) && textIdx < len(text) {
		queryChar := query[queryIdx]

		found := false
		for textIdx < len(text) {
			if text[textIdx] == queryChar {
				found = true

				positionScore := len(text) - textIdx

				if textIdx == lastMatchIdx+1 {
					consecutiveMatches++
					positionScore += consecutiveMatches * 10
				} else {
					consecutiveMatches = 0
				}

				if textIdx == 0 || isWordBoundary(text[textIdx-1]) {
					positionScore += 15
				}

				if textIdx > 0 && isLower(rune(text[textIdx-1])) && isUpper(rune(text[textIdx])) {
					positionScore += 10
				}

				score += positionScore
				lastMatchIdx = textIdx
				queryIdx++
				textIdx++
				break
			}
			textIdx++
		}

		if !found {
			return 0
		}
	}

	if queryIdx == len(query) {
		if len(text) < 50 {
			score += (50 - len(text)) * 2
		}
		return score
	}

	return 0
}

func (f *FuzzyMatcher) sortByScore(matches []ScoredItem) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Score > matches[i].Score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

func isWordBoundary(c byte) bool {
	return c == ' ' || c == '-' || c == '_' || c == '.' || c == '/' || c == '\\'
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}
