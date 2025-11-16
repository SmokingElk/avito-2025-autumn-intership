package dto

import "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/statistics/entity"

type AssignmentsPerMember struct {
	MemberId         string `db:"member_id"`
	AssignmentsCount int    `db:"assigments_count"`
}

func (a *AssignmentsPerMember) ToAssignmentsPerMemberEntity() entity.AssignmentsPerMember {
	return entity.AssignmentsPerMember{
		MemberId:         a.MemberId,
		AssignmentsCount: a.AssignmentsCount,
	}
}
