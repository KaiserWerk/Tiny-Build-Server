package main

import (
	"errors"
)

func getUserByEmail(n string) (user, error) {
	db, err := getDbConnection()
	if err != nil {
		return user{}, errors.New("could not get database connection")
	}

	row := db.QueryRow("SELECT id, displayname, email, password, locked FROM user WHERE email = ?", n)
	var u user
	err = row.Scan(&u.id, &u.displayname, &u.email, &u.password, &u.locked)
	if err != nil {
		return user{}, errors.New("could not scan")
	}

	return u, nil
}
