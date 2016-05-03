package envdb

import "errors"

// Query database table.
type QueryDb struct {
	Id    int64  `json:"id"`
	Name  string `xorm:"NOT NULL"`
	Query string `xorm:"NOT NULL"`
	Type  string `xorm:"NOT NULL"`
}

// Load all of the default saved queries. This will
// only run once on initial database creation.
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

	q5 := QueryDb{
		Name:  "Third-party kernel extensions (OS X)",
		Query: "SELECT * FROM kernel_extensions WHERE name NOT LIKE 'com.apple.%' AND name != '__kernel__';",
		Type:  "all",
	}

	queries = append(queries, q5)

	q6 := QueryDb{
		Name:  "Startup items (OS X / LaunchDaemons & LaunchAgents)",
		Query: `SELECT disabled, path, program FROM launchd;`,
		Type:  "all",
	}

	queries = append(queries, q6)

	q7 := QueryDb{
		Name:  "Shell history",
		Query: `SELECT * FROM shell_history;`,
		Type:  "all",
	}

	queries = append(queries, q7)

	q8 := QueryDb{
		Name:  "All users with group information",
		Query: `SELECT * FROM users u JOIN groups g where u.gid = g.gid;`,
		Type:  "all",
	}

	queries = append(queries, q8)

	q9 := QueryDb{
		Name: "Interface information",
		Query: `SELECT address, mac, id.interface
FROM interface_details AS id, interface_addresses AS ia WHERE id.interface = ia.interface;`,
		Type: "all",
	}

	queries = append(queries, q9)

	// q8 := QueryDb{
	// Name: "All empty groups",
	// Query: `SELECT groups.gid, groups.name FROM groups
	// LEFT JOIN users ON (groups.gid = users.gid) WHERE users.uid IS NULL;`,
	// Type: "all",
	// }

	// queries = append(queries, q8)

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

// Find a saved query by its id.
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

// Delete a saved query
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

// Find all saved queries in the database.
func AllSavedQueries() ([]*QueryDb, error) {
	var data []*QueryDb
	err := x.Find(&data)

	return data, err
}

// Insert a new saved query to the database.
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
