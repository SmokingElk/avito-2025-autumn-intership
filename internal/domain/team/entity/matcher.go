package entity

import (
	"fmt"
)

type TeamMatcher struct {
	expected Team
}

func (m *TeamMatcher) Matches(x interface{}) bool {
	actual, ok := x.(Team)
	if !ok {
		return false
	}

	if m.expected.Name != actual.Name {
		return false
	}

	if len(m.expected.Members) != len(actual.Members) {
		return false
	}

	for i, member := range actual.Members {
		if m.expected.Members[i].Id != member.Id {
			return false
		}
	}

	return true
}

func (m *TeamMatcher) String() string {
	return fmt.Sprintf("matches team %+v", m.expected)
}

func Matcher(team Team) *TeamMatcher {
	return &TeamMatcher{
		expected: team,
	}
}
