package dto

import "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"

type MemberDTO struct {
	Id       string  `db:"id"`
	Activity string  `db:"activity"`
	Name     string  `db:"username"`
	TeamName *string `db:"team_name"`
}

func (m MemberDTO) ToMemberEntity() entity.Member {
	teamName := "no team"
	if m.TeamName != nil {
		teamName = *m.TeamName
	}

	return entity.Member{
		Id:       m.Id,
		Activity: entity.MemberActivity(m.Activity),
		Username: m.Name,
		TeamName: teamName,
	}
}
