package zookeeper

type MonitorBridge interface {
	RegisterMember(memberId string, memberInfo string) error
	CheckMemberDead(memberId string)  error
	//MemberIsStarting(memberId string, memberInfo string) error
}