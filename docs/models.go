package docs

import (
	memberEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/member/entity"
	prEntity "github.com/SmokingElk/avito-2025-autumn-intership/internal/domain/pull-request/entity"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewErrorResponse(code, message string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
	}
}

type SetIsActiveRequest struct {
	UserId   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

type SetIsActiveResponse struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

func ToSetIsActiveResponse(member memberEntity.Member) SetIsActiveResponse {
	return SetIsActiveResponse{
		UserId:   member.Id,
		Username: member.Username,
		TeamName: member.TeamName,
		IsActive: member.Activity == memberEntity.MemberActive,
	}
}

type GetReviewPRResponse struct {
	Id       string `json:"pull_request_id"`
	Name     string `json:"pull_request_name"`
	AuthorId string `json:"author_id"`
	Status   string `json:"status"`
}

func ToReviewPRResponse(pr prEntity.PullRequest) GetReviewPRResponse {
	return GetReviewPRResponse{
		Id:       pr.Id,
		Name:     pr.Name,
		AuthorId: pr.AuthorId,
		Status:   string(pr.Status),
	}
}

type GetReviewResponse struct {
	UserId       string                `json:"user_id"`
	PullRequests []GetReviewPRResponse `json:"pull_requests"`
}

type TeamMember struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

func ToTeamMember(member memberEntity.Member) TeamMember {
	return TeamMember{
		UserId:   member.Id,
		Username: member.Username,
		IsActive: member.Activity == memberEntity.MemberActive,
	}
}

func (m *TeamMember) ToTeamMemberEntity() memberEntity.Member {
	activity := memberEntity.MemberActive
	if !m.IsActive {
		activity = memberEntity.MemberInactive
	}

	return memberEntity.Member{
		Id:       m.UserId,
		Username: m.Username,
		Activity: activity,
	}
}

type AddTeamRequest struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type AddTeamResponseObject struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}

type AddTeamResponse struct {
	Team AddTeamResponseObject `json:"team"`
}

type GetTeamResponse struct {
	Name    string       `json:"team_name"`
	Members []TeamMember `json:"members"`
}
