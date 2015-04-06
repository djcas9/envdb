package main

import "errors"

type QueryDb struct {
	Id    int64  `json:"id"`
	Name  string `xorm:"NOT NULL"`
	Query string `xorm:"NOT NULL"`
	Type  string `xorm:"NOT NULL"`
}

func LoadDefaultSavedQueries() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	queries := []QueryDb{}

	q1 := QueryDb{
		Name:  "Process List",
		Query: "select * from processes;",
		Type:  "all",
	}

	queries = append(queries, q1)

	q2 := QueryDb{
		Name:  "All listening ports joined with processes",
		Query: "select * from listening_ports a join processes b on a.pid = b.pid;",
		Type:  "all",
	}

	queries = append(queries, q2)

	q3 := QueryDb{
		Name:  "All currently executing processes where the original binary no longer exists",
		Query: "SELECT name, path, pid FROM processes WHERE on_disk = 0;",
		Type:  "all",
	}

	queries = append(queries, q3)

	q4 := QueryDb{
		Name: "All processes that are listening on network ports",
		Query: `SELECT DISTINCT process.name, listening.port, listening.address, process.pid
FROM processes AS process 
JOIN listening_ports 
AS listening ON process.pid = listening.pid;`,
		Type: "all",
	}

	queries = append(queries, q4)

	if _, err := sess.Insert(&queries); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}

func FindSavedQueryById(id int64) (*QueryDb, error) {
	Log.Debugf("Looking for saved query with id: %d", id)

	query := &QueryDb{Id: id}

	has, err := x.Get(query)

	if err != nil {
		return nil, err
	} else if !has {
		return nil, errors.New("Saved Query not found")
	}

	return query, nil
}

func (self *QueryDb) Delete() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Delete(&QueryDb{Id: self.Id}); err != nil {
		Log.Debug("Saved Query Delete Error: ", err)
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}

func AllSavedQueries() ([]*QueryDb, error) {
	var data []*QueryDb
	err := x.Find(&data)

	return data, err
}

func NewSavedQuery(self QueryDb) error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Insert(self); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}
