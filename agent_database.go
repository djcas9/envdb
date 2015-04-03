package main

import (
	"errors"
	"time"
)

type DbAgent struct {
	Id int64

	AgentId  string
	Name     string
	Ip       string
	Hostname string

	Online bool

	OsQuery        bool
	OsQueryVersion string

	Created time.Time `xorm:"CREATED"`
	Updated time.Time `xorm:"UPDATED"`
}

func AllAgents() ([]*DbAgent, error) {
	var agents []*DbAgent
	err := x.Find(&agents)

	return agents, err
}

func (self *DbAgent) Update() error {
	_, err := x.Id(self.Id).AllCols().Update(self)
	return err
}

func AgentUpdateOrCreate(agent *AgentData) (*DbAgent, error) {
	find, err := GetAgentByAgentId(agent.Id)

	if find != nil {
		Log.Debug("Found existing node record.")

		find.Name = agent.Name
		find.Ip = agent.Ip
		find.Hostname = agent.Hostname
		find.OsQuery = agent.OsQuery
		find.OsQueryVersion = agent.OsQueryVersion
		find.Online = agent.Online

		Log.Debug("Found Agent: ", find.Online, agent.Online)

		_, err := x.Id(find.Id).AllCols().Update(find)

		if err != nil {
			return find, err
		}

		return find, nil
	}

	Log.Debugf("Couldn't find record for node (%s).", agent.Id)
	Log.Debugf("Creating a new record.")

	a := &DbAgent{
		AgentId:        agent.Id,
		Name:           agent.Name,
		Ip:             agent.Ip,
		Hostname:       agent.Hostname,
		Online:         agent.Online,
		OsQuery:        agent.OsQuery,
		OsQueryVersion: agent.OsQueryVersion,
	}

	_, err = x.Insert(a)

	if err != nil {
		return nil, err
	}

	return a, nil
}

func GetAgentByAgentId(agentId string) (*DbAgent, error) {
	agent := &DbAgent{AgentId: agentId}

	has, err := x.Get(agent)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Agent not found")
	}

	return agent, nil
}
