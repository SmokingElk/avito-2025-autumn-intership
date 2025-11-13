package dto

import (
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	teamEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/team/entity"
)

type MemberDTO struct {
	Id       string  `db:"id"`
	Username string  `db:"username"`
	Activity string  `db:"activity"`
	TeamId   *string `db:"team_id"`
}

func (m MemberDTO) ToMemberEntity() memberEntity.Member {
	return memberEntity.Member{
		Id:       m.Id,
		Username: m.Username,
		Activity: memberEntity.MemberActivity(m.Activity),
		TeamId:   m.TeamId,
	}
}

type TeamDTO struct {
	Id      string `db:"id"`
	Name    string `db:"name"`
	Members []MemberDTO
}

func (t TeamDTO) ToTeamEntity() teamEntity.Team {
	members := make([]memberEntity.Member, 0, len(t.Members))

	for _, member := range t.Members {
		members = append(members, member.ToMemberEntity())
	}

	return teamEntity.Team{
		Id:      t.Id,
		Name:    t.Name,
		Members: members,
	}
}
