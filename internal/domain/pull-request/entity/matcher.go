package entity

import (
	"fmt"
)

type PRMatcher struct {
	expected PullRequest
}

func (m *PRMatcher) Matches(x interface{}) bool {
	actual, ok := x.(PullRequest)
	if !ok {
		return false
	}

	if m.expected.Name != actual.Name ||
		m.expected.Id != actual.Id ||
		m.expected.AuthorId != actual.AuthorId ||
		m.expected.Status != actual.Status {

		return false
	}

	if len(m.expected.Reviewers) != len(actual.Reviewers) {
		return false
	}

	reviewersMap := make(map[string]struct{})

	for _, reviewer := range m.expected.Reviewers {
		reviewersMap[reviewer] = struct{}{}
	}

	for _, reviewer := range actual.Reviewers {
		if _, ok := reviewersMap[reviewer]; !ok {
			return false
		}
	}

	return true
}

func (m *PRMatcher) String() string {
	return fmt.Sprintf("matches pull request %+v", m.expected)
}

func Matcher(pr PullRequest) *PRMatcher {
	return &PRMatcher{
		expected: pr,
	}
}
