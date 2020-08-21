package main

import (
	"errors"
)

func getUserByEmail(n string) (user, error) {
	db, err := getDbConnection()
	if err != nil {
		return user{}, errors.New("could not get database connection")
	}

	row := db.QueryRow("SELECT Id, Displayname, Email, Password, Locked, Admin FROM user WHERE Email = ?", n)
	var u user
	//var Locked int
	//var Admin int
	err = row.Scan(&u.Id, &u.Displayname, &u.Email, &u.Password, &u.Locked, &u.Admin)
	if err != nil {
		return user{}, errors.New("could not scan")
	}

	return u, nil
}
