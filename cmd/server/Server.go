package server

import (
	"Cataloguer/cmd/cache"
	"Cataloguer/cmd/cache/redistore"
	"Cataloguer/cmd/model"
	"Cataloguer/cmd/store"
	"Cataloguer/cmd/store/sqlstore"
	"Cataloguer/cmd/util"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

//TODO LIST:
//1) Тесты
//2) Конфиг
//3) Логи
//4) Многопоточность

type Server struct {
	Router           *mux.Router
	Context          context.Context
	Store            store.Store
	Cache            cache.Cache
	FunctionalServer http.Server
	//Config
	//Auth
}

type ctxKey int

const (
	sessionID               = "session_id"
	cookieExpiration        = 3600 * 24
	ctxKeyUser       ctxKey = iota
)

// Config
func (s *Server) Start() {
	/*
		http.Server{
			Addr:              "",
			Handler:           nil,
			ReadTimeout:       0,
			ReadHeaderTimeout: 0,
			WriteTimeout:      0,
		}

	*/

	s.ConfigureRoutes()
	err := http.ListenAndServe("127.0.0.1:8080", s)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.Router.ServeHTTP(w, r)
}

func New() *Server {
	// todo добавить config
	redisCache, err1 := redistore.New("someurl")
	sqlStore, err2 := sqlstore.New("someurl")
	if err1 != nil || err2 != nil {
		log.Fatal(err1.Error(), err2.Error())
		return nil
	}
	return &Server{
		Router: mux.NewRouter(),
		Store:  sqlStore,
		Cache:  redisCache,
		//Config:
	}
}

func (s *Server) ConfigureRoutes() {
	s.Router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"http://127.0.0.1:5500"}), handlers.AllowCredentials()))
	s.Router.HandleFunc("/register", s.register).Methods("POST")
	s.Router.HandleFunc("/login", s.logIn()).Methods("POST")
	privateAccess := s.Router.PathPrefix("/account").Subrouter()
	privateAccess.Use(s.auth)
	privateAccess.HandleFunc("/message", uploading).Methods("POST")
}

// middleware для проверки валидности session_id
func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionID)
		if err != nil {
			if cookie != nil {
				a := s.Cache.Session()
				a.DeleteValue(cookie.Value)
			}
			sendError(w, r, http.StatusForbidden, http.ErrNoCookie)
			return
		}

		userID, err := s.Cache.Session().GetValue(cookie.Value)
		if err != nil {
			sendError(w, r, http.StatusForbidden, errors.New("You must be logged in"))
			return
		}
		id, _ := strconv.Atoi(userID)
		u, err := s.Store.User().FindByID(id)
		if err != nil {
			s.Cache.Session().DeleteValue(cookie.Value)
			sendError(w, r, http.StatusForbidden, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	u := &model.User{}
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		sendError(w, r, http.StatusBadRequest, err)
		return
	}
	if err = u.ValidateUserFields(); err != nil {
		sendError(w, r, http.StatusBadRequest, err)
		return
	}
	err = s.Store.User().SaveUser(u)
	if err != nil {
		sendError(w, r, http.StatusConflict, errors.New("User with this email already registered"))
		return
	}
	s.createSession(w, r, u)
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request, u *model.User) {
	//todo заменить на средство понадежнее
	sessionId := util.RandString(64)
	err := s.Cache.Session().SetValue(sessionId, strconv.Itoa(u.ID), cookieExpiration)
	if err != nil {
		sendError(w, r, http.StatusInternalServerError, err)
		return
	}

	cookie := http.Cookie{Name: sessionID, Value: sessionId, Expires: time.Now().Add(time.Second * cookieExpiration), Path: "/", HttpOnly: true}
	//w.Header().Add("Access-Control-Allow-Credentials", "true")
	http.SetCookie(w, &cookie)

	w.WriteHeader(http.StatusOK)
}

// todo создать свой тип ошибок
func (s *Server) logIn() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	//отправлять что пользователь с такой почтой уже зарегистрирован.
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			sendError(w, r, http.StatusBadRequest, err)
			return
		}
		u, err := s.Store.User().FindByEmail(req.Email)
		if err != nil {
			sendError(w, r, http.StatusUnauthorized, errors.New("The user with this email doesn't exist"))
			return
		}
		if u.IsPasswordCorrect(req.Password) {
			s.createSession(w, r, u)
			return
		}
		sendError(w, r, http.StatusUnauthorized, errors.New("Wrong password or email"))
	}
}

func sendError(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	resp := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(resp)
}

func uploading(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	err := r.ParseMultipartForm(32 << 20) // upload of 32 MB files.
	file, handler, err := r.FormFile("item")
	if err != nil {
		sendError(w, r, http.StatusBadRequest, errors.New("Error Retrieving the File, There's no 'item' tag in formdata.\n"+err.Error()))
	}
	defer file.Close()

	tmp1 := handler.Header.Get("Content-Type")
	name := tmp1[:strings.LastIndex(tmp1, "/")]

	tmp2 := handler.Filename
	tp := tmp2[strings.LastIndex(tmp2, ".")+1:]
	tempFile, err := os.CreateTemp("data/user1", name+"*."+tp)
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	tempFile.Write(fileBytes)
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func (s *Server) chat(w http.ResponseWriter, r *http.Request) {

}
