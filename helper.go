package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
)

func writeToConsole(s string) {
	fmt.Println("> " + s)
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func getDbConnection() (*sql.DB, error) {
	db, err := sql.Open(centralConfig.Database.Driver, centralConfig.Database.DSN)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getConfiguration() {
	cont, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic("Could not read configuration file '"+configFile+"': " + err.Error())
	}

	err = yaml.Unmarshal(cont, &centralConfig)
	if err != nil {
		panic("could not parse configuration file content: " + err.Error())
	}
}

func generateToken(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hex.EncodeToString(b)
}

func hashString(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func doesHashMatch(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}



