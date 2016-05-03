package envdb

import (
	"fmt"
	"os"
	"path"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	// _ "github.com/cznic/ql/driver"
	// _ "github.com/go-xorm/ql"
)

var (
	x      *xorm.Engine
	tables []interface{}
)

// DBInit will initialize the database and add default values.
func DBInit(storePath, logPath string) error {
	tables = append(tables, new(NodeDb), new(QueryDb), new(SettingsDb),
		new(UserDb))

	if err := SetEngine(storePath, logPath); err != nil {
		return err
	}

	s, err := DbSettings()

	if err != nil {
		Log.Error("Unable to create database settings.")
		return err
	}

	if !s.Setup {

		user := &UserDb{
			Name:     "Administrator",
			Email:    "admin@envdb.io",
			Password: "envdb",
		}

		if err := CreateUser(user); err != nil {
			Log.Fatal("Unable to create default admin user.")
		}

		if err := LoadDefaultSavedQueries(); err != nil {
			Log.Error("Unable to load default saved queries.")
			Log.Errorf("Error: %s", err)
		}

		s.Setup = true

		if err := s.Update(); err != nil {
			Log.Errorf("Unable to update settings: %s", err)
		}
	}

	return nil
}

func getEngine(DbPath string) (*xorm.Engine, error) {
	cnnstr := ""
	os.MkdirAll(path.Dir(DbPath), os.ModePerm)
	cnnstr = "file:" + DbPath + "?cache=shared&mode=rwc"
	// cnnstr = DbPath

	return xorm.NewEngine("sqlite3", cnnstr)
}

// SetEngine will setup and connect to the database.
func SetEngine(DbPath string, LogPath string) (err error) {
	x, err = getEngine(DbPath)

	if err != nil {
		return fmt.Errorf("connect to database: %v", err)
	}

	logPath := path.Join(LogPath, "critical-stack-intel-sql.log")

	os.MkdirAll(path.Dir(logPath), os.ModePerm)

	f, err := os.Create(logPath)

	if err != nil {
		return fmt.Errorf("models.init(fail to create critical-stack-intel-sql.log): %v", err)
	}

	x.SetLogger(xorm.NewSimpleLogger(f))

	x.ShowSQL = false
	x.ShowInfo = false
	x.ShowDebug = false
	x.ShowErr = false
	x.ShowWarn = false

	if err = x.StoreEngine("InnoDB").Sync2(tables...); err != nil {
		return fmt.Errorf("sync database struct error: %v", err)
	}

	return nil
}
