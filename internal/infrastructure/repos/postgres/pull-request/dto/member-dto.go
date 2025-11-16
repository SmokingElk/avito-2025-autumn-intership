package dto

import "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"

type MemberDTO struct {
	Id       string `db:"id"`
	Activity string `db:"activity"`
}

func (m MemberDTO) ToMemberEntity() entity.Member {
	return entity.Member{
		Id:       m.Id,
		Activity: entity.MemberActivity(m.Activity),
	}
}
