package main

import (
	"errors"
	"time"
)

type NodeDb struct {
	Id int64

	NodeId   string
	Name     string
	Ip       string
	Hostname string
	Os       string

	Online bool

	OsQuery        bool
	OsQueryVersion string

	PendingDelete bool

	Created time.Time `xorm:"CREATED"`
	Updated time.Time `xorm:"UPDATED"`
}

func AllNodes() ([]*NodeDb, error) {
	var nodes []*NodeDb
	err := x.Find(&nodes)

	return nodes, err
}

func (self *NodeDb) Update() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(self.Id).AllCols().Update(self); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return err
}

func NodeUpdateOrCreate(node *NodeData) (*NodeDb, error) {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return nil, err
	}

	find, err := GetNodeByNodeId(node.Id)

	if find != nil {
		Log.Debug("Found existing node record.")

		find.Name = node.Name
		find.Ip = node.Ip
		find.Hostname = node.Hostname
		find.Os = node.Os
		find.OsQuery = node.OsQuery
		find.OsQueryVersion = node.OsQueryVersion
		find.Online = node.Online
		find.PendingDelete = node.PendingDelete

		if _, err := sess.Id(find.Id).AllCols().Update(find); err != nil {
			sess.Rollback()
			return find, err
		}

		err := sess.Commit()

		if err != nil {
			return nil, err
		}

		return find, nil
	}

	Log.Debugf("Error: %s", err)

	Log.Debugf("Creating a new record.")

	a := &NodeDb{
		NodeId:         node.Id,
		Name:           node.Name,
		Ip:             node.Ip,
		Hostname:       node.Hostname,
		Os:             node.Os,
		Online:         node.Online,
		OsQuery:        node.OsQuery,
		OsQueryVersion: node.OsQueryVersion,
		PendingDelete:  false,
	}

	if _, err := sess.Insert(a); err != nil {
		sess.Rollback()
		return nil, err
	}

	err = sess.Commit()

	if err != nil {
		return nil, err
	}

	return a, nil
}

func GetNodeByNodeId(nodeId string) (*NodeDb, error) {
	Log.Debugf("Looking for node with id: %s", nodeId)

	node := &NodeDb{NodeId: nodeId}

	has, err := x.Get(node)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Node not found")
	}

	return node, nil
}

func (self *NodeDb) Delete() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(self.Id).Delete(self); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}
