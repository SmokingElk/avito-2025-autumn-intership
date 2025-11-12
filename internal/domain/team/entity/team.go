package entity

import (
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	"github.com/google/uuid"
)

type Team struct {
	Id      string
	Name    string
	Members []memberEntity.Member
}

func NewTeam(name string, members []memberEntity.Member) Team {
	id := uuid.NewString()

	return Team{
		Id:      id,
		Name:    name,
		Members: members,
	}
}
