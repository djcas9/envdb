package envdb

import "testing"

func TestServer_Database(t *testing.T) {
	if x == nil {
		t.Fatal("No database engine found. Sqlite error.")
	}

	if x.DriverName() != "sqlite3" {
		t.Fatal("Database is using the wrong engine.")
	}
}

func TestServer_Database_DefaultUser(t *testing.T) {
	user, err := FindUserByEmail("admin@envdb.io")

	if err != nil {
		t.Fatal("NewServer should create a default user.")
	}

	if !user.Admin {
		t.Fatal("User should has admin rights.")
	}

	if !user.ValidatePassword("envdb") {
		t.Fatal("Default user has incorrect password.")
	}
}

func TestServer_Database_FindAllUsers(t *testing.T) {
	users, err := FindAllUsers()

	if err != nil {
		t.Fatal(err)
	}

	if len(users) < 1 {
		t.Fatal("No users found")
	}
}
