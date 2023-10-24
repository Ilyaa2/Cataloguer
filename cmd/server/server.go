package server

import (
	"Cataloguer/cmd/cache"
	"Cataloguer/cmd/cache/redistore"
	"Cataloguer/cmd/custom_errors"
	"Cataloguer/cmd/model"
	"Cataloguer/cmd/store"
	"Cataloguer/cmd/store/sqlstore"
	"Cataloguer/cmd/util"
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Context          context.Context
	Store            store.Store
	Cache            cache.Cache
	FunctionalServer *http.Server
	Auth             Auth
	Config           Config
}

type ctxKey int

const (
	CtxKeyUser ctxKey = iota
)

func (s *Server) Start() {
	s.Auth = NewAuth(s)
	//mux.NewRouter()
	s.FunctionalServer.Handler = s.ConfigureRoutes()
	log.Fatal(s.FunctionalServer.ListenAndServe())
}

func New(config Config) *Server {
	redisCache, err1 := redistore.New(config.CacheUrl)
	sqlStore, err2 := sqlstore.New(config.StoreUrl)
	if err1 != nil || err2 != nil {
		log.Fatal(err1.Error(), err2.Error())
		return nil
	}
	return &Server{
		Store:  sqlStore,
		Cache:  redisCache,
		Config: config,
		FunctionalServer: &http.Server{
			Addr:         config.ServerAddr,
			Handler:      nil,
			ReadTimeout:  time.Hour * 10,
			WriteTimeout: time.Hour * 10,
		},
	}
}

func (s *Server) ConfigureRoutes() http.Handler {
	router := mux.NewRouter()
	router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"http://127.0.0.1:5500"}), handlers.AllowCredentials()))
	router.HandleFunc("/register", s.Auth.Register).Methods("POST")
	router.HandleFunc("/login", s.Auth.Login).Methods("POST")

	privateAccess := router.PathPrefix("/account").Subrouter()
	privateAccess.Use(s.Auth.Auth)

	privateAccess.HandleFunc("/message", s.uploadMessage).Methods("POST")
	privateAccess.HandleFunc("/message_id", s.deleteMessageId).Methods("DELETE")
	privateAccess.HandleFunc("/message_name", s.deleteMessageName).Methods("DELETE")
	privateAccess.HandleFunc("/messages_list", s.messagesList).Methods("GET")
	privateAccess.PathPrefix("/messages/").HandlerFunc(s.messages).Methods("GET")
	return router
}

func (s *Server) deleteMessageId(w http.ResponseWriter, r *http.Request) {
	req := model.MessageId{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectJsonStructure))
		return
	}
	if !s.Store.Message().HasRightsOnMessageById(getUserIDFromContext(r), req.Id) {
		SendError(w, r, http.StatusForbidden, errors.New(custom_errors.NotEnoughRights))
		return
	}

	msg, err := s.Store.Message().GetMessageById(req.Id)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.Store.Message().DeleteMessageById(req.Id)
	deleteMessage(msg, s.Config.BasePath)
	w.WriteHeader(http.StatusOK)
}

func deleteMessage(msg *model.Message, basePath string) {
	relFilePath := util.StripUrlFromFilePath(msg.Path, "/account")
	absFilePath := filepath.ToSlash(filepath.Join(basePath, relFilePath))
	err := os.Remove(absFilePath)
	_ = err
}

func (s *Server) deleteMessageName(w http.ResponseWriter, r *http.Request) {
	req := model.MessageName{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectJsonStructure))
		return
	}
	if !s.Store.Message().HasRightsOnMessageByName(getUserIDFromContext(r), req.Name) {
		SendError(w, r, http.StatusForbidden, errors.New(custom_errors.NotEnoughRights))
		return
	}

	msg, err := s.Store.Message().GetMessageByName(req.Name)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	s.Store.Message().DeleteMessageByName(req.Name)
	deleteMessage(msg, s.Config.BasePath)
	w.WriteHeader(http.StatusOK)
}

func doHaveRights(r *http.Request) bool {
	userId := getUserIDFromContext(r)
	requestedId, err := util.NumberOfUsersDirectory(r.URL.Path)
	if err != nil || requestedId != userId {
		return false
	}
	return true
}

func (s *Server) messages(w http.ResponseWriter, r *http.Request) {
	if !doHaveRights(r) {
		SendError(w, r, http.StatusForbidden, errors.New(custom_errors.NotEnoughRights))
		return
	}
	fs := http.FileServer(http.Dir(filepath.Join(s.Config.BasePath, "./messages")))
	http.StripPrefix("/account/messages/", fs).ServeHTTP(w, r)
}

func SendError(w http.ResponseWriter, r *http.Request, code int, err error) {
	w.WriteHeader(code)
	resp := map[string]string{"error": err.Error()}
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) uploadMessage(w http.ResponseWriter, r *http.Request) {
	log.Println("File Upload Endpoint Hit")
	//time.Sleep(time.Second * 10)
	err := r.ParseMultipartForm(1 << 20) // 0 - might help.
	defer r.MultipartForm.RemoveAll()

	formFile, handler, err := r.FormFile("item")
	tp := r.FormValue("type")
	if err != nil || tp == "" {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.IncorrectMultipartFile+err.Error()))
		return
	}
	userId := getUserIDFromContext(r)
	filePath, err := s.createFileForUser(formFile, handler, userId)
	if err != nil {
		log.Println(err)
		SendError(w, r, http.StatusInternalServerError, errors.New(custom_errors.ServerSide))
		return
	}
	err = s.createMessageInDB(userId, filePath, tp)
	if err != nil {
		log.Println(err)
		SendError(w, r, http.StatusInternalServerError, errors.New(custom_errors.ServerSide))
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"detailed_response": "Successfully Uploaded File\n"})
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) createMessageInDB(userId int, filePath string, fileType string) error {
	message := &model.Message{
		Type:     fileType,
		Name:     util.StripFileNameFromPath(filePath),
		DateTime: time.Now(),
		Path:     "/account" + util.StripUrlFromFilePath(filePath, s.Config.BasePath),
	}
	err := s.Store.Message().CreateMessage(message, userId)
	return err
}

func (s *Server) createFileForUser(formFile multipart.File, handler *multipart.FileHeader, userId int) (string, error) {
	tmp1 := handler.Header.Get("Content-Type")
	name := tmp1[:strings.LastIndex(tmp1, "/")]

	tmp2 := handler.Filename
	ext := tmp2[strings.LastIndex(tmp2, ".")+1:]

	dirPath := filepath.ToSlash(filepath.Join(s.Config.BasePath, "messages/", "user"+strconv.Itoa(userId)))
	err := os.Mkdir(dirPath, os.ModeDir)
	if err != nil && !os.IsExist(err) {
		return "", err
	}

	tempFile, err := os.CreateTemp(dirPath, name+"*."+ext)
	if err != nil {
		return "", err
	}

	//_, err = io.CopyBuffer(tempFile, formFile, make([]byte, 1024*64))
	_, err = io.Copy(tempFile, formFile)
	tempFile.Close()
	formFile.Close()
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(tempFile.Name()), nil
}

func getUserIDFromContext(r *http.Request) int {
	u := r.Context().Value(CtxKeyUser).(*model.User)
	return u.ID
}

func (s *Server) messagesList(w http.ResponseWriter, r *http.Request) {
	messages, err := s.Store.Message().GetMessagesByUserId(getUserIDFromContext(r))
	if err != nil {
		SendError(w, r, http.StatusBadRequest, errors.New(custom_errors.NoFilesFound))
	}
	response := model.MessagesResponse{Messages: messages, BaseUrl: s.Config.BaseUrl}
	json.NewEncoder(w).Encode(response)
}
