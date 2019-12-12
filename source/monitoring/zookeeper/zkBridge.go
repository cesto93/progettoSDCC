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
	IsDead bool
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

func New(zkServerAddresses []string, sessionTimeout time.Duration, aliveNodePath string) (*ZookeeperBridge, error) {
	//Connect to servers	
	var bridge ZookeeperBridge 
	var err error
	var conn *zk.Conn	
	conn, _, err = zk.Connect(zkServerAddresses, sessionTimeout)
	if err != nil {
		return &bridge, fmt.Errorf("Error in zkBridge Conn: %v", err)
	}
	bridge = ZookeeperBridge{conn, sessionTimeout, aliveNodePath, false}

	//Create root for membership if not exist
	if checkAndCreateNode(aliveNodePath, bridge.zkConn) != nil {
		return &bridge, err
	}
	return &bridge, nil
}

func (bridge *ZookeeperBridge) RegisterMember(memberId string, memberInfo string) error {
	//Save memberinfo in /membershipNodepath/memberId
	pathAlive := fmt.Sprintf("%s/%s", bridge.aliveNodePath, memberId)
	_, err := bridge.zkConn.Create(pathAlive, []byte(memberInfo), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		return fmt.Errorf("Error in zkBridge RegisterMember Create: %v", err)
	}
	return nil
}

func (bridge *ZookeeperBridge) CheckMemberDead(memberId string)  error {
	var alive bool
	path := fmt.Sprintf("%s/%s", bridge.aliveNodePath, memberId)

	//Wait changes in memebership nodes
	_, _, watcher, err := bridge.zkConn.ExistsW(path)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge ExistW: %v", err)
	}
	<-watcher
	alive,_,err = bridge.zkConn.Exists(path)
	if  err != nil {
		return fmt.Errorf("Error in zkBridge Esist: %v", err)
	}
	bridge.IsDead = !alive
	return nil
}
