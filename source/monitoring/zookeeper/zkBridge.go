package zookeeper

import (
	"time"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type ZookeeperBridge struct {
	zkConn *zk.Conn
	sessionTimeout time.Duration
	membershipNodePath string
	Members []string
	MembersDead []string
}

func New(zkServerAddresses []string, sessionTimeout time.Duration, 
			membershipNodePath string, members []string) (*ZookeeperBridge, error) {
	//Connect to servers	
	var bridge ZookeeperBridge 
	var err error
	var conn *zk.Conn
	var nodeExist bool		
	conn, _, err = zk.Connect(zkServerAddresses, sessionTimeout)
	if err != nil {
		return &bridge, fmt.Errorf("Error in zkBridge Conn: %v", err)
	}
	bridge = ZookeeperBridge{conn, sessionTimeout, membershipNodePath, members, nil}

	//Create root for membership if not exist
	nodeExist, _, err = bridge.zkConn.Exists(membershipNodePath);
	if err != nil {
		return &bridge, fmt.Errorf("Error in zkBridge Exist: %v", err)
	}
	if !nodeExist {
		_, err = bridge.zkConn.Create(membershipNodePath, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return &bridge, fmt.Errorf("Error in zkBridge Create: %v", err)
		}
	}
	return &bridge, nil
}

func (bridge *ZookeeperBridge) RegisterMember(memberId string, memberInfo string) error {
	//Save memberinfo in /membershipNodepath/memberId
	path := fmt.Sprintf("%s/%s", bridge.membershipNodePath, memberId)
	_, err := bridge.zkConn.Create(path, []byte(memberInfo), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("Error in zkBridge RegisterMember Create: %v", err)
	}
	return nil
}


func (bridge *ZookeeperBridge) CheckMembersDead()  error {
	//Wait changes in memebership nodes
	_, _, watcher, err := bridge.zkConn.ChildrenW(bridge.membershipNodePath)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge ChildrenW: %v", err)
	}
	<-watcher

	//Updates nodes dead
	nodesAlive, _, err := bridge.zkConn.Children(bridge.membershipNodePath)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge Children: %v", err)
	}
	bridge.MembersDead = getMembersDead(nodesAlive, bridge.Members)
	return nil
}

func getMembersDead(nodesAlive []string, nodes []string) []string {
	var dead []string 
	var i,j int
	for i, _ = range  nodes {
		for j, _ = range nodesAlive {
			if nodes[i] == nodesAlive[j] {
				break
			}
		}
		if j == len(nodesAlive) {
			dead = append(dead, nodes[i])
		}
	}
	return dead
}