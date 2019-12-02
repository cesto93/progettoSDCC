package zookeeper

import (
	"time"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
)

type ZookeeperBridge struct {
	zkConn *zk.Conn
	sessionTimeout time.Duration
	aliveNodePath string
	startingNodePath string
	Members []string
	//MembersStarting []string
	MembersDead []string
}

func checkAndCreateNode(path string, conn *zk.Conn) error {
	nodeExist, _, err := conn.Exists(path);
	if err != nil {
		return fmt.Errorf("Error in zkBridge Exist: %v", err)
	}
	if !nodeExist {
		_, err = conn.Create(path, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return fmt.Errorf("Error in zkBridge Create: %v", err)
		}
	}
	return nil
}

func checkAndDeleteNode(path string, conn *zk.Conn) error {
	nodeExist, _, err := conn.Exists(path);
	if err != nil {
		return fmt.Errorf("Error in zkBridge Exist: %v", err)
	}

	if !nodeExist {
		err = conn.Delete(path, -1)
		if err != nil {
			return fmt.Errorf("Error in zkBridge RegisterMember Delete: %v", err)
		}
	}
	return nil
}

func New(zkServerAddresses []string, sessionTimeout time.Duration, 
			aliveNodePath string, startingNodePath string , members []string) (*ZookeeperBridge, error) {
	//Connect to servers	
	var bridge ZookeeperBridge 
	var err error
	var conn *zk.Conn	
	conn, _, err = zk.Connect(zkServerAddresses, sessionTimeout)
	if err != nil {
		return &bridge, fmt.Errorf("Error in zkBridge Conn: %v", err)
	}
	bridge = ZookeeperBridge{conn, sessionTimeout, aliveNodePath, startingNodePath, members, nil}

	//Create root for membership if not exist
	if checkAndCreateNode(aliveNodePath, bridge.zkConn) != nil {
		return &bridge, err
	}

	if checkAndCreateNode(startingNodePath, bridge.zkConn) != nil {
		return &bridge, err
	}

	/*nodeExist, _, err = bridge.zkConn.Exists(AliveNodePath);
	if err != nil {
		return &bridge, fmt.Errorf("Error in zkBridge Exist: %v", err)
	}
	if !nodeExist {
		_, err = bridge.zkConn.Create(AliveNodePath, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			return &bridge, fmt.Errorf("Error in zkBridge Create: %v", err)
		}
	}*/
	return &bridge, nil
}

func (bridge *ZookeeperBridge) RegisterMember(memberId string, memberInfo string) error {
	//Save memberinfo in /membershipNodepath/memberId
	pathAlive := fmt.Sprintf("%s/%s", bridge.aliveNodePath, memberId)
	pathStarting := fmt.Sprintf("%s/%s", bridge.startingNodePath, memberId)
	_, err := bridge.zkConn.Create(pathAlive, []byte(memberInfo), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("Error in zkBridge RegisterMember Create: %v", err)
	}
	checkAndDeleteNode(pathStarting, bridge.zkConn)
	return nil
}

func (bridge *ZookeeperBridge) MemberIsStarting(memberId string, memberInfo string) error {
	//Save memberinfo in /membershipNodepath/memberId
	path := fmt.Sprintf("%s/%s", bridge.startingNodePath, memberId)
	_, err := bridge.zkConn.Create(path, []byte(memberInfo), 0, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("Error in zkBridge RegisterMember Create: %v", err)
	}
	return nil
}


func (bridge *ZookeeperBridge) CheckMembersDead()  error {
	//Wait changes in memebership nodes
	_, _, watcher, err := bridge.zkConn.ChildrenW(bridge.aliveNodePath)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge ChildrenW: %v", err)
	}
	<-watcher

	//Updates nodes dead
	nodesAlive, _, err := bridge.zkConn.Children(bridge.aliveNodePath)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge Children: %v", err)
	}
	nodesStarting, _, err := bridge.zkConn.Children(bridge.startingNodePath)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge Children: %v", err)
	}
	bridge.MembersDead = getMembersDead(nodesAlive, nodesStarting, bridge.Members)
	return nil
}

func getMembersDead(nodesAlive []string, nodesStarting []string, nodes []string) []string {
	var dead []string 
	var i,j int
	nodesAlive = append(nodesAlive, nodesStarting...)
	for i, _ = range  nodes {
		for j = 0; j < len(nodesAlive); j++ { //DON'T SUBSTITUTE WITH RANGE OR IT WILL NOT WORK
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