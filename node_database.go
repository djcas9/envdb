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

	Online bool

	OsQuery        bool
	OsQueryVersion string

	Created time.Time `xorm:"CREATED"`
	Updated time.Time `xorm:"UPDATED"`
}

func AllNodes() ([]*NodeDb, error) {
	var nodes []*NodeDb
	err := x.Find(&nodes)

	return nodes, err
}

func (self *NodeDb) Update() error {
	_, err := x.Id(self.Id).AllCols().Update(self)
	return err
}

func NodeUpdateOrCreate(node *NodeData) (*NodeDb, error) {
	find, err := GetNodeByNodeId(node.Id)

	if find != nil {
		Log.Debug("Found existing node record.")

		find.Name = node.Name
		find.Ip = node.Ip
		find.Hostname = node.Hostname
		find.OsQuery = node.OsQuery
		find.OsQueryVersion = node.OsQueryVersion
		find.Online = node.Online

		Log.Debug("Found Node: ", find.Online, node.Online)

		_, err := x.Id(find.Id).AllCols().Update(find)

		if err != nil {
			return find, err
		}

		return find, nil
	}

	Log.Debugf("Couldn't find record for node (%s).", node.Id)
	Log.Debugf("Error: %s", err)

	Log.Debugf("Creating a new record.")

	a := &NodeDb{
		NodeId:         node.Id,
		Name:           node.Name,
		Ip:             node.Ip,
		Hostname:       node.Hostname,
		Online:         node.Online,
		OsQuery:        node.OsQuery,
		OsQueryVersion: node.OsQueryVersion,
	}

	_, err = x.Insert(a)

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
