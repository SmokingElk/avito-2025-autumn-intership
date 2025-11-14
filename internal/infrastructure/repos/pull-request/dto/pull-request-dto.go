package dto

import (
	"time"

	"github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
	"github.com/lib/pq"
)

type PullRequestDTO struct {
	Id        string         `db:"id"`
	Name      string         `db:"pr_name"`
	AuthorId  string         `db:"author_id"`
	Status    string         `db:"pr_status"`
	CreatedAt time.Time      `db:"created_at"`
	TeamId    string         `db:"team_id"`
	MergedAt  time.Time      `db:"merged_at"`
	Reviewers pq.StringArray `db:"reviewers"`
}

func (pr PullRequestDTO) ToPullRequestEntity() entity.PullRequest {
	return entity.PullRequest{
		Id:        pr.Id,
		Name:      pr.Name,
		AuthorId:  pr.AuthorId,
		Status:    entity.PRStatus(pr.Status),
		CreatedAt: pr.CreatedAt,
		MergedAt:  pr.MergedAt,
		Reviewers: pr.Reviewers,
	}
}
