package main

type SettingsDb struct {
	Id               int64
	LoadSavedQueries bool
}

func DbSettings() (*SettingsDb, error) {
	s := &SettingsDb{Id: 1}

	Log.Debugf("Looking for %s database settings.", Name)
	has, err := x.Get(s)

	if err != nil {
		return nil, err
	} else if !has {
		Log.Debug("Couldn't find settings. Creating a new settings row.")

		if _, err := x.Insert(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

func (self *SettingsDb) Update() error {
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

	return nil
}
