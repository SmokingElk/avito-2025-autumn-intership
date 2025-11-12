package entity

type MemberActivity string

const (
	MemberActive   MemberActivity = "ACTIVE"
	MemberInactive MemberActivity = "INACTIVE"
)

type Member struct {
	Id       string
	Username string
	Activity MemberActivity
	TeamId   *string
}
