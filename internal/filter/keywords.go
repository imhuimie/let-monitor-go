// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Keyword Filter Implementation"
//   Timestamp: "2025-11-25T13:45:10Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python keyword filtering from core.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "AND/OR keyword matching logic implemented"
// }}

package filter

import (
	"strings"
)

// KeywordFilter filters content based on keyword rules
type KeywordFilter struct {
	rule string
}

// NewKeywordFilter creates a new keyword filter
func NewKeywordFilter(rule string) *KeywordFilter {
	return &KeywordFilter{rule: rule}
}

// Match checks if text matches the keyword rule
// Rule format: "keyword1+keyword2,keyword3" means (keyword1 AND keyword2) OR keyword3
func (f *KeywordFilter) Match(text string) bool {
	if f.rule == "" {
		return false
	}

	text = strings.ToLower(text)
	orGroups := strings.Split(f.rule, ",")

	for _, group := range orGroups {
		group = strings.TrimSpace(group)
		andKeywords := strings.Split(group, "+")

		allMatch := true
		for _, keyword := range andKeywords {
			keyword = strings.TrimSpace(strings.ToLower(keyword))
			if !strings.Contains(text, keyword) {
				allMatch = false
				break
			}
		}

		if allMatch {
			return true
		}
	}

	return false
}
