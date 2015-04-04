package main

import (
	"fmt"
	"os"
	"path"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	x       *xorm.Engine
	tables  []interface{}
	API_KEY string = ""
)

func DBInit(storePath, logPath string) error {
	tables = append(tables, new(NodeDb), new(QueryDb), new(SettingsDb))

	if err := SetEngine(storePath, logPath); err != nil {
		return err
	}

	s, err := DbSettings()

	if err != nil {
		Log.Error("Unable to create database settings.")
		return err
	}

	Log.Debug(s)

	if !s.LoadSavedQueries {
		if err := LoadDefaultSavedQueries(); err != nil {
			Log.Error("Unable to load default saved queries.")
			Log.Errorf("Error: %s", err)
		}

		s.LoadSavedQueries = true

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

	return xorm.NewEngine("sqlite3", cnnstr)
}

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
		return fmt.Errorf("sync database struct error: %v\n", err)
	}

	return nil
}
