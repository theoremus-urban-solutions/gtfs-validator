package parser

import "strings"

// IDInterner keeps one copy of each ID, so a stored ID does not hold its whole CSV line in memory.
type IDInterner map[string]string

// NewIDInterner returns an empty interner.
func NewIDInterner() IDInterner {
	return make(IDInterner)
}

// Intern returns the stored copy of id, making one on first sight.
func (ids IDInterner) Intern(id string) string {
	if stored, ok := ids[id]; ok {
		return stored
	}
	stored := strings.Clone(id)
	ids[stored] = stored
	return stored
}
