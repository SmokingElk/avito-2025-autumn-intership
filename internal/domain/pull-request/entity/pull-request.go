package entity

import "time"

type PRStatus string

const (
	PROpen   PRStatus = "OPEN"
	PRMerged PRStatus = "MERGED"
)

type PullRequest struct {
	Id        string
	Name      string
	AuthorId  string
	Status    PRStatus
	CreatedAt time.Time
	MergedAt  time.Time
	Reviewers []string
}

func NewPullRequest(id, name, authorId string) PullRequest {
	return PullRequest{
		Id:        id,
		Name:      name,
		AuthorId:  authorId,
		Status:    PROpen,
		CreatedAt: time.Now(),
	}
}
