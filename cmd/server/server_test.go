package server

import (
	"Cataloguer/cmd/model"
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type TransferThings struct {
	cookies     []*http.Cookie
	server      *Server
	resp        *httptest.ResponseRecorder
	buf         *bytes.Buffer
	req         *http.Request
	allMessages *model.MessagesResponse
}

func registerOrLogin1(t *testing.T, url string) []*http.Cookie {
	payload := map[string]string{
		"email":    "conic@gmail.com",
		"password": "123123",
		"name":     "Lopic",
	}
	server := prepareServer()
	resp := httptest.NewRecorder()
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(payload)
	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		t.Fatal(err)
	}
	server.FunctionalServer.Handler.ServeHTTP(resp, req)
	return resp.Result().Cookies()
}

func registerOrLogin2(t *testing.T, url string) []*http.Cookie {
	payload := map[string]string{
		"email":    "ponip@gmail.com",
		"password": "123123",
		"name":     "zdarova",
	}
	server := prepareServer()
	resp := httptest.NewRecorder()
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(payload)
	req, err := http.NewRequest(http.MethodPost, url, b)
	if err != nil {
		t.Fatal(err)
	}
	server.FunctionalServer.Handler.ServeHTTP(resp, req)
	return resp.Result().Cookies()
}

func credentials1(t *testing.T) []*http.Cookie {
	cookies := registerOrLogin1(t, "/register")
	if len(cookies) == 0 {
		return registerOrLogin1(t, "/login")
	}
	return cookies
}

func credentials2(t *testing.T) []*http.Cookie {
	cookies := registerOrLogin2(t, "/register")
	if len(cookies) == 0 {
		return registerOrLogin2(t, "/login")
	}
	return cookies
}

func TestCreateMessage(t *testing.T) {
	cookies := credentials1(t)
	server := prepareServer()
	values := []map[string]io.Reader{
		{
			"item": mustOpen(filepath.Join(server.Config.BasePath, "test_resources/to_upload/video8492912843.mp4")),
			"type": strings.NewReader("video"),
		},
		{
			"item": mustOpen(filepath.Join(server.Config.BasePath, "test_resources/to_upload/text7483827581.txt")),
			"type": strings.NewReader("text"),
		},
		{
			"item": mustOpen(filepath.Join(server.Config.BasePath, "test_resources/to_upload/image573728403.png")),
			"type": strings.NewReader("image"),
		},
		{
			"item": mustOpen(filepath.Join(server.Config.BasePath, "test_resources/to_upload/audio1738248599.mp3")),
			"type": strings.NewReader("audio"),
		},
	}
	for _, val := range values {
		resp := httptest.NewRecorder()
		multipartWriter, buf := upload(t, val)

		req, err := http.NewRequest(http.MethodPost, "/account/message", buf)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(cookies[0])
		req.AddCookie(cookies[1])
		req.Header.Set("Content-Type", multipartWriter.FormDataContentType())
		server.FunctionalServer.Handler.ServeHTTP(resp, req)
		val["item"].(io.Closer).Close()
		assert.Equal(t, http.StatusOK, resp.Code)
	}
}

func getAllMessages(t *testing.T) TransferThings {
	cookies := credentials1(t)
	server := prepareServer()
	resp := httptest.NewRecorder()
	b := bytes.Buffer{}
	req, err := http.NewRequest(http.MethodGet, "/account/messages_list", &b)
	if err != nil {
		t.Fatal(err)
	}
	req.AddCookie(cookies[0])
	req.AddCookie(cookies[1])
	server.FunctionalServer.Handler.ServeHTTP(resp, req)

	var allMessages model.MessagesResponse
	json.NewDecoder(resp.Body).Decode(&allMessages)
	return TransferThings{
		cookies:     cookies,
		server:      server,
		buf:         &b,
		allMessages: &allMessages,
		resp:        resp,
	}
}

func TestGetAllMessages(t *testing.T) {
	tt := getAllMessages(t)
	for _, msg := range tt.allMessages.Messages {
		t.Log(msg.Id, msg.Name, msg.Type, msg.Path, msg.DateTime)
		tt.buf.Reset()
		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, msg.Path, tt.buf)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(tt.cookies[0])
		req.AddCookie(tt.cookies[1])
		tt.server.FunctionalServer.Handler.ServeHTTP(resp, req)
		_, filename := filepath.Split(msg.Path)
		pathFile := filepath.Join(tt.server.Config.BasePath, "test_resources/to_download/", filename)
		file, err := os.Create(filepath.ToSlash(pathFile))
		if err != nil {
			t.Fatal(err)
		}
		io.Copy(file, resp.Body)
		file.Close()
		assert.FileExists(t, pathFile)
	}
}

func TestAddressToForeignFiles(t *testing.T) {
	tt := getAllMessages(t)
	tt.cookies = credentials2(t)
	for _, msg := range tt.allMessages.Messages {
		t.Log(msg.Id, msg.Name, msg.Type, msg.Path, msg.DateTime)
		tt.buf.Reset()
		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodGet, msg.Path, tt.buf)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(tt.cookies[0])
		req.AddCookie(tt.cookies[1])
		tt.server.FunctionalServer.Handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusForbidden, resp.Code)
	}
}

func TestDeleteOwnFilesById(t *testing.T) {
	TestCreateMessage(t)
	TestCreateMessage(t)
	TestCreateMessage(t)
	tt := getAllMessages(t)
	for _, msg := range tt.allMessages.Messages {
		t.Log(msg.Id, msg.Name, msg.Type, msg.Path, msg.DateTime)
		tt.buf.Reset()
		json.NewEncoder(tt.buf).Encode(map[string]int{
			"message_id": msg.Id,
		})
		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodDelete, "/account/message_id", tt.buf)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(tt.cookies[0])
		req.AddCookie(tt.cookies[1])
		tt.server.FunctionalServer.Handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}
	tt = getAllMessages(t)
	assert.Equal(t, http.StatusBadRequest, tt.resp.Code)
}

func TestDeleteOwnFilesByName(t *testing.T) {
	TestCreateMessage(t)
	TestCreateMessage(t)
	TestCreateMessage(t)
	tt := getAllMessages(t)
	for _, msg := range tt.allMessages.Messages {
		t.Log(msg.Id, msg.Name, msg.Type, msg.Path, msg.DateTime)
		tt.buf.Reset()
		json.NewEncoder(tt.buf).Encode(map[string]string{
			"message_name": msg.Name,
		})
		resp := httptest.NewRecorder()
		req, err := http.NewRequest(http.MethodDelete, "/account/message_name", tt.buf)
		if err != nil {
			t.Fatal(err)
		}
		req.AddCookie(tt.cookies[0])
		req.AddCookie(tt.cookies[1])
		tt.server.FunctionalServer.Handler.ServeHTTP(resp, req)
		assert.Equal(t, http.StatusOK, resp.Code)
	}
	tt = getAllMessages(t)
	assert.Equal(t, http.StatusBadRequest, tt.resp.Code)
}

func mustOpen(path string) *os.File {
	fileToUpload, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	return fileToUpload
}

func upload(t *testing.T, values map[string]io.Reader) (*multipart.Writer, *bytes.Buffer) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	var err error
	for key, reader := range values {
		var fw io.Writer
		if closer, ok := reader.(io.Closer); ok {
			defer closer.Close()
		}
		if file, ok := reader.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, file.Name()); err != nil {
				t.Fatal(err)
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				t.Fatal(err)
			}
		}
		if _, err = io.Copy(fw, reader); err != nil {
			t.Fatal(err)
		}
	}

	w.Close()
	return w, &b
}
