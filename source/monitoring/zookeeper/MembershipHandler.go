package zookeeper

type MonitorBridge interface {
	RegisterMember(memberId string, memberInfo string) error
	CheckMembers()  error
	MemberIsStarting(memberId string, memberInfo string) error
}