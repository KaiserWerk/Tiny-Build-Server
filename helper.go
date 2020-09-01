package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	"os"
	sessionstore "session-store"
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

func getConfiguration() configuration {
	cont, err := ioutil.ReadFile(configFile)
	if err != nil {
		panic("Could not read configuration file '"+configFile+"': " + err.Error())
	}

	var cfg configuration
	err = yaml.Unmarshal(cont, &cfg)
	if err != nil {
		panic("could not parse configuration file content: " + err.Error())
	}

	return cfg
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

func sendMail(m *gomail.Message) {
	// fetch data from system configuration
	d := gomail.NewDialer("smtp.example.com", 587, "user", "123456")
	if err := d.DialAndSend(m); err != nil {
		writeToConsole("could not send email: " + err.Error())
	}
}

func checkLogin(r *http.Request) (sessionstore.Session, error) {
	sessId, err := sessMgr.GetCookieValue(r)
	if err != nil {
		return sessionstore.Session{}, errors.New("could not get cookie: " + err.Error())
	}
	writeToConsole("getting session with Id "+sessId)
	session, err := sessMgr.GetSession(sessId)
	if err != nil {
		return sessionstore.Session{}, errors.New("could not get session: " + err.Error())
	}

	return session, nil
}


