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
	Context          context.Context
	Store            store.Store
	Cache            cache.Cache
	FunctionalServer *http.Server
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
	s.FunctionalServer.Handler = s.ConfigureRoutes()
	log.Fatal(s.FunctionalServer.ListenAndServe())
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
		Store: sqlStore,
		Cache: redisCache,
		//Config:
		FunctionalServer: &http.Server{
			Addr:         "127.0.0.1:8080",
			Handler:      nil,
			ReadTimeout:  time.Hour * 10,
			WriteTimeout: time.Hour * 10,
		},
	}
}

func (s *Server) ConfigureRoutes() http.Handler {
	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"http://127.0.0.1:5500"}), handlers.AllowCredentials()))
	router.HandleFunc("/register", s.register).Methods("POST")
	router.HandleFunc("/login", s.logIn()).Methods("POST")
	privateAccess := router.PathPrefix("/account").Subrouter()
	privateAccess.Use(s.auth)
	privateAccess.HandleFunc("/message", uploading).Methods("POST")
	return router
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
			sendMessage(w, r, http.StatusForbidden, http.ErrNoCookie)
			return
		}

		userID, err := s.Cache.Session().GetValue(cookie.Value)
		if err != nil {
			sendMessage(w, r, http.StatusForbidden, errors.New("You must be logged in"))
			return
		}
		id, _ := strconv.Atoi(userID)
		u, err := s.Store.User().FindByID(id)
		if err != nil {
			s.Cache.Session().DeleteValue(cookie.Value)
			sendMessage(w, r, http.StatusForbidden, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
		//_ = u
		//next.ServeHTTP(w, r)
	})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	u := &model.User{}
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		fmt.Println(err.Error())
		sendMessage(w, r, http.StatusBadRequest, err)
		return
	}
	if err = u.ValidateUserFields(); err != nil {
		sendMessage(w, r, http.StatusBadRequest, err)
		return
	}
	err = s.Store.User().SaveUser(u)
	if err != nil {
		sendMessage(w, r, http.StatusConflict, errors.New("User with this email already registered"))
		return
	}
	s.createSession(w, r, u)
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request, u *model.User) {
	//todo заменить на средство понадежнее
	sessionId := util.RandString(64)
	err := s.Cache.Session().SetValue(sessionId, strconv.Itoa(u.ID), cookieExpiration)
	if err != nil {
		sendMessage(w, r, http.StatusInternalServerError, err)
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

	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			sendMessage(w, r, http.StatusBadRequest, err)
			return
		}
		u, err := s.Store.User().FindByEmail(req.Email)
		if err != nil {
			sendMessage(w, r, http.StatusUnauthorized, errors.New("The user with this email doesn't exist"))
			return
		}
		if u.IsPasswordCorrect(req.Password) {
			s.createSession(w, r, u)
			return
		}
		sendMessage(w, r, http.StatusUnauthorized, errors.New("Wrong password or email"))
	}
}

func sendMessage(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	resp := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(resp)
}

// https://freshman.tech/file-upload-golang/
// прогресс бар на сервере.
func uploading(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")
	time.Sleep(time.Second * 10)
	err := r.ParseMultipartForm(1 << 20) // 0 - might help.
	defer func() {
		if err := r.MultipartForm.RemoveAll(); err != nil {
			log.Printf("failed to free multipart form resources: %v", err)
		}
	}()

	formFile, handler, err := r.FormFile("item")

	if err != nil {
		sendMessage(w, r, http.StatusBadRequest, errors.New("Error Retrieving the File, There's no 'item' tag in formdata.\n"+err.Error()))
	}

	tmp1 := handler.Header.Get("Content-Type")
	name := tmp1[:strings.LastIndex(tmp1, "/")]

	tmp2 := handler.Filename
	tp := tmp2[strings.LastIndex(tmp2, ".")+1:]

	dirName := "user" + getIDFromContext(r)
	err = os.Mkdir("data/"+dirName, os.ModeDir)
	if err != nil && !os.IsExist(err) {
		sendMessage(w, r, http.StatusInternalServerError, err)
		return
	}

	tempFile, err := os.CreateTemp("data/"+dirName, name+"*."+tp)
	if err != nil {
		fmt.Println(err)
		sendMessage(w, r, http.StatusInternalServerError, err)
		return
	}

	//_, err = io.CopyBuffer(tempFile, formFile, make([]byte, 1024*64))
	_, err = io.Copy(tempFile, formFile)
	tempFile.Close()
	formFile.Close()
	if err != nil {
		fmt.Println(err)
		sendMessage(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"detailed_response": "Successfully Uploaded File\n"})
}

func getIDFromContext(r *http.Request) string {
	u := r.Context().Value(ctxKeyUser).(*model.User)
	return strconv.Itoa(u.ID)
}

/*
func (s *Server) chat(w http.ResponseWriter, r *http.Request) {

}
*/
