package server

import (
	"Cataloguer/cmd/custom_errors"
	"Cataloguer/cmd/model"
	"Cataloguer/cmd/util"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

type CookieAuth struct {
	Server *Server
}

const (
	sessionID        = "session_id"
	UserName         = "user_name"
	cookieExpiration = 3600 * 24
)

func NewAuth(s *Server) Auth {
	return &CookieAuth{Server: s}
}

func (c *CookieAuth) Login(w http.ResponseWriter, r *http.Request) {
	req := &model.User{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || req.IsPayloadFieldsEmpty() {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectJsonStructure))
		return
	}
	u, err := c.Server.Store.User().FindByEmail(req.Email)
	if err != nil {
		SendError(w, r, http.StatusUnauthorized, errors.New(custom_errors.EmailNotRegistered))
		return
	}
	if u.IsPasswordCorrect(req.Password) {
		c.createSession(w, r, u)
		w.WriteHeader(http.StatusOK)
		return
	}
	SendError(w, r, http.StatusUnauthorized, errors.New(custom_errors.WrongPasswordOrEmail))
}

func (c *CookieAuth) Register(w http.ResponseWriter, r *http.Request) {
	u := &model.User{}
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectJsonStructure))
		return
	}
	if err = u.ValidateUserFields(); err != nil {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectUsersFields+err.Error()))
		return
	}
	err = c.Server.Store.User().Save(u)
	if err != nil {
		SendError(w, r, http.StatusConflict, errors.New(custom_errors.ThisEmailAlreadyRegistered))
		return
	}
	c.createSession(w, r, u)
	w.WriteHeader(http.StatusCreated)
}

func (c *CookieAuth) createSession(w http.ResponseWriter, r *http.Request, u *model.User) {
	sessionId := util.RandString(64)
	err := c.Server.Cache.Session().SetValue(sessionId, strconv.Itoa(u.ID), cookieExpiration)
	if err != nil {
		log.Println(err)
		SendError(w, r, http.StatusInternalServerError, errors.New(custom_errors.ServerSide))
		return
	}

	cookie1 := http.Cookie{
		Name:     sessionID,
		Value:    sessionId,
		Expires:  time.Now().Add(time.Second * cookieExpiration),
		Path:     "/",
		HttpOnly: true,
	}
	http.SetCookie(w, &cookie1)

	cookie2 := http.Cookie{
		Name:    UserName,
		Value:   u.Name,
		Expires: time.Now().Add(time.Second * cookieExpiration),
		Path:    "/",
	}
	http.SetCookie(w, &cookie2)
}

// Auth - middleware for checking whether session_id is valid
func (c *CookieAuth) Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionID)
		if err != nil {
			if cookie != nil {
				a := c.Server.Cache.Session()
				a.DeleteValue(cookie.Value)
			}
			SendError(w, r, http.StatusForbidden, http.ErrNoCookie)
			return
		}

		userID, err := c.Server.Cache.Session().GetValue(cookie.Value)
		if err != nil {
			SendError(w, r, http.StatusForbidden, errors.New(custom_errors.UserDidntLogIn))
			return
		}
		id, _ := strconv.Atoi(userID)
		u, err := c.Server.Store.User().FindByID(id)
		if err != nil {
			c.Server.Cache.Session().DeleteValue(cookie.Value)
			SendError(w, r, http.StatusForbidden, errors.New(custom_errors.UserDidntLogIn))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxKeyUser, u)))
	})
}
