package envdb

// Settings database table
type SettingsDb struct {
	Id    int64
	Setup bool
}

// Initialize the database settings
func DbSettings() (*SettingsDb, error) {
	s := &SettingsDb{Id: 1}

	Log.Debugf("Looking for %s database settings.", Name)
	has, err := x.Get(s)

	if err != nil {
		return nil, err
	} else if !has {
		Log.Debug("Couldn't find settings. Creating a new settings row.")

		sess := x.NewSession()
		defer sess.Close()

		if err := sess.Begin(); err != nil {
			return nil, err
		}

		if _, err := sess.Insert(s); err != nil {
			sess.Rollback()
			return nil, err
		}

		err := sess.Commit()

		if err != nil {
			return nil, err
		}
	}

	return s, nil
}

// Update settings in the database
func (s *SettingsDb) Update() error {
	sess := x.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Id(s.Id).AllCols().Update(s); err != nil {
		sess.Rollback()
		return err
	}

	err := sess.Commit()

	if err != nil {
		return err
	}

	return nil
}
