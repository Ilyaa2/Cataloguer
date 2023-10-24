package server

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieAuth_Register(t *testing.T) {
	server := prepareServer()
	testCases := []struct {
		testName     string
		payload      map[string]string
		expectedCode int
	}{
		{
			testName: "valid",
			payload: map[string]string{
				"email":    "user@example.org",
				"password": "password",
				"name":     "valid",
			},
			expectedCode: http.StatusCreated,
		},
		{
			testName: "invalid password",
			payload: map[string]string{
				"email":    "user@example.org",
				"password": "p",
				"name":     "smth",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName: "invalid structure",
			payload: map[string]string{
				"password": "password",
				"name":     "valfeid",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName: "invalid email",
			payload: map[string]string{
				"email":    "userexa.org",
				"password": "password",
				"name":     "valid",
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			testName: "user already exists",
			payload: map[string]string{
				"email":    "user@example.org",
				"password": "password",
				"name":     "valid",
			},
			expectedCode: http.StatusConflict,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			resp := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(testCase.payload)
			req, _ := http.NewRequest(http.MethodPost, "/register", b)
			server.FunctionalServer.Handler.ServeHTTP(resp, req)
			assert.Equal(t, testCase.expectedCode, resp.Code)
		})
	}
}

func TestCookieAuth_Login(t *testing.T) {
	server := prepareServer()
	testCases := []struct {
		testName     string
		payload      map[string]string
		expectedCode int
	}{
		{
			testName: "valid",
			payload: map[string]string{
				"email":    "user@example.org",
				"password": "password",
			},
			expectedCode: http.StatusOK,
		},
		{
			testName: "invalid password",
			payload: map[string]string{
				"email":    "user@example.org",
				"password": "invalid password",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			testName: "email not registered",
			payload: map[string]string{
				"email":    "newEmail@example.org",
				"password": "passw",
			},
			expectedCode: http.StatusUnauthorized,
		},
		{
			testName: "invalid structure",
			payload: map[string]string{
				"email": "user@example.org",
				"pass":  "pppp",
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			resp := httptest.NewRecorder()
			b := &bytes.Buffer{}
			json.NewEncoder(b).Encode(testCase.payload)
			req, _ := http.NewRequest(http.MethodPost, "/login", b)
			server.FunctionalServer.Handler.ServeHTTP(resp, req)
			assert.Equal(t, testCase.expectedCode, resp.Code)
		})
	}
}

func TestCookieAuth_Auth(t *testing.T) {
	server := prepareServer()
	payload := map[string]string{
		"email":    "user@example.org",
		"password": "password",
	}
	resp := httptest.NewRecorder()
	b := &bytes.Buffer{}
	json.NewEncoder(b).Encode(payload)
	req, _ := http.NewRequest(http.MethodPost, "/login", b)
	server.FunctionalServer.Handler.ServeHTTP(resp, req)
	cookies := resp.Result().Cookies()

	cookieSessionId := cookies[0]
	cookieName := cookies[1]
	t.Log(cookieName.Name + " : " + cookieName.Value)
	t.Log(cookieSessionId.Name + " : " + cookieSessionId.Value)

	req, _ = http.NewRequest(http.MethodGet, "/account/messages_list", b)
	req.AddCookie(cookieName)
	req.AddCookie(cookieSessionId)
	assert.NotEqual(t, http.StatusForbidden, resp.Code)
}
