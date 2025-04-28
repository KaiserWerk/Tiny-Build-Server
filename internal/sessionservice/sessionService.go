package sessionservice

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/KaiserWerk/Tiny-Build-Server/internal/dbservice"
	"github.com/KaiserWerk/Tiny-Build-Server/internal/entity"

	"github.com/KaiserWerk/sessionstore/v2"
)

type Session = sessionstore.Session

type ISessionService interface {
	AddMessage(w http.ResponseWriter, t, msg string)
	SetMessage(t string, content string)
	GetMessage(w http.ResponseWriter, r *http.Request) (string, string, error)
	CreateSession(lt time.Time) (*Session, error)
	GetSession(id string) (*Session, error)
	GetSessionFromCookie(r *http.Request) (*Session, error)
	RemoveSession(id string) error
	RemoveAllSessions()
	SetCookie(w http.ResponseWriter, value string, expires time.Time)
	RemoveCookie(w http.ResponseWriter, name string)
	GetCookieValue(r *http.Request) (string, error)
}

type SessionService struct {
	sessMgr *sessionstore.SessionManager
}

// NewSessionService creates a new session manager
func NewSessionService(name string) ISessionService {
	mgr := sessionstore.NewManager(name)
	return &SessionService{
		sessMgr: mgr,
	}
}

func (m *SessionService) AddMessage(w http.ResponseWriter, t, msg string) {
	m.sessMgr.AddMessage(w, t, msg)
}

func (m *SessionService) SetMessage(t string, content string) {
	//m.sessMgr.SetMessage(t, content) // TODO: fix??
}

func (m *SessionService) GetMessage(w http.ResponseWriter, r *http.Request) (string, string, error) {
	return m.sessMgr.GetMessage(w, r)
}

func (m *SessionService) CreateSession(lt time.Time) (*Session, error) {
	return m.sessMgr.CreateSession(lt)
}

func (m *SessionService) GetSession(id string) (*Session, error) {
	return m.sessMgr.GetSession(id)
}

func (m *SessionService) GetSessionFromCookie(r *http.Request) (*Session, error) {
	return m.sessMgr.GetSessionFromCookie(r)
}

func (m *SessionService) RemoveSession(id string) error {
	return m.sessMgr.RemoveSession(id)
}

func (m *SessionService) RemoveAllSessions() {
	m.sessMgr.RemoveAllSessions()
}

func (m *SessionService) SetCookie(w http.ResponseWriter, value string, expires time.Time) {
	m.sessMgr.SetCookie(w, value, expires)
}

func (m *SessionService) RemoveCookie(w http.ResponseWriter, name string) {
	m.sessMgr.RemoveCookie(w, name)
}

func (m *SessionService) GetCookieValue(r *http.Request) (string, error) {
	return m.sessMgr.GetCookieValue(r)
}

// GetUserFromSession returns a user from a given session, if possible
func GetUserFromSession(ds dbservice.IDBService, s *sessionstore.Session) (entity.User, error) {
	userIdStr, ok := s.GetVar("user_id")
	if !ok {
		return entity.User{}, fmt.Errorf("variable key user_id not found")
	}

	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		return entity.User{}, err
	}

	user, err := ds.GetUserById(uint(userId))
	return user, err
}
