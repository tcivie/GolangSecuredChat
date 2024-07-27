package db

import "testing"

func TestOpenUsersDB(t *testing.T) {
	if db, err := OpenUsersDB(); err != nil {
		t.Errorf("Error opening users database: %v", err)
	} else {
		defer db.conn.Close()
	}
}
