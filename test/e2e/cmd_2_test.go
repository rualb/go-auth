package e2e

import (
	"bytes"
	"encoding/json"
	"go-auth/internal/config/consts"
	"go-auth/internal/controller"
	"go-auth/internal/service"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	auth "go-auth/internal/controller/auth"
	auth_tel "go-auth/internal/controller/auth/tel"
	xweb "go-auth/internal/web"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAuthEchoHandler(t *testing.T) {

	type StatusResponse struct {
		IsAuth bool `json:"is_auth"`
	}

	{

		e := echo.New()

		// Define test cases
		tests := []struct {
			title    string
			cookie   string
			expected StatusResponse
		}{
			{
				title:    "test 1",
				cookie:   "_auth=" + url.QueryEscape(getAuthToken(appService, "auth", false)),
				expected: StatusResponse{IsAuth: true},
			},
			{
				title:    "test 2",
				cookie:   "_auth=" + url.QueryEscape(getAuthToken(appService, "not-auth", false)),
				expected: StatusResponse{IsAuth: false},
			},
			{
				title:    "test 3",
				cookie:   "",
				expected: StatusResponse{IsAuth: false},
			},
		}

		for _, tc := range tests {
			t.Run(tc.title, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, consts.PathAuthStatusAPI, nil)
				rec := httptest.NewRecorder()

				if tc.cookie != "" {
					req.Header.Set("Cookie", tc.cookie)
				}

				c := e.NewContext(req, rec)

				// Setup handler with middleware
				handler := func(c echo.Context) error {
					ctrl := auth.NewStatusAPIController(appService, c)
					return ctrl.Handler()
				}

				handler = xweb.TokenParserMiddleware(appService)(handler)
				handler = xweb.UserLangMiddleware(appService)(handler)

				if err := handler(c); err != nil {
					t.Errorf("Handler error: %v", err)
				}

				var response StatusResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if !reflect.DeepEqual(response, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, response)
				}
			})
		}
	}
}

func TestAuthSigninEchoHandler(t *testing.T) {

	/*
		const testOnlyID = "test-only-1"
		const testOnlyTel = "+999999999999"
		const testOnlyPw = "123456"
	*/

	type statusResponse struct {
		Status string `json:"status,omitempty"`
		// IsSuccess bool   `json:"is_success"`
	}
	type reqBody struct {
		Tel      string `json:"tel"`
		Password string `json:"password"`
	}
	{

		e := echo.New()

		// Setup handler with middleware
		handler := func(c echo.Context) error {
			ctrl := auth_tel.NewAccountSigninController(appService, c)
			return ctrl.Handler()
		}

		handler = xweb.TokenParserMiddleware(appService)(handler)
		handler = xweb.UserLangMiddleware(appService)(handler)

		req := httptest.NewRequest(http.MethodGet, consts.PathAuthSigninTelAPI, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		if err := handler(c); err != nil {
			t.Fatalf("Handler error: %v", err)
		}
		// csrfToken := rec.Header().Get("X-Csrf-Token") // case sensitive
		// if csrfToken == "" {
		// 	t.Fatalf("x-csrf-token is empty")
		// }
		// Define test cases
		tests := []struct {
			title    string
			cookie   string
			body     any
			expected statusResponse
		}{
			{
				title:    "test 1",
				cookie:   "", // "_csrf=" + url.QueryEscape(csrfToken),
				expected: statusResponse{Status: "success"},
				body:     &reqBody{Tel: testOnlyUserTel, Password: testOnlyUserPw},
			},
			{
				title:    "test 2",
				cookie:   "", // "_csrf=" + url.QueryEscape(csrfToken),
				expected: statusResponse{Status: ""},
				body:     &reqBody{Tel: testOnlyUserTel, Password: testOnlyUserPw + "1"},
			},
			{
				title:    "test 3",
				cookie:   "", // "_csrf=" + url.QueryEscape(csrfToken),
				expected: statusResponse{Status: ""},
				body:     &reqBody{Tel: testOnlyUserTel + "1", Password: testOnlyUserPw},
			},
			{
				title:    "test 1-2",
				cookie:   "", // "_csrf=" + url.QueryEscape(csrfToken),
				expected: statusResponse{Status: "success"},
				body:     &reqBody{Tel: testOnlyUserTel, Password: testOnlyUserPw},
			},
		}

		for _, tc := range tests {
			t.Run(tc.title, func(t *testing.T) {

				body, _ := json.Marshal(tc.body)

				req := httptest.NewRequest(http.MethodPost, consts.PathAuthSigninTelAPI,
					bytes.NewReader(body),
				)

				rec := httptest.NewRecorder()

				if tc.cookie != "" {
					req.Header.Set("Cookie", tc.cookie)
				}

				req.Header.Set(`Content-Type`, "application/json")

				c := e.NewContext(req, rec)

				if err := handler(c); err != nil {
					t.Errorf("Handler error: %v", err)
				}

				var response statusResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if !reflect.DeepEqual(response, tc.expected) {
					t.Errorf("Expected %+v, got %+v", tc.expected, response)
				}
			})
		}
	}
}

func TestAuthSigninRawEchoHandler(t *testing.T) {

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, consts.PathAuthSigninTelAPI, nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)
	signInService := controller.SignInService(c, appService)

	// resp.Header() has +1 _auth cookie

	accountService := appService.Account()
	user, err := accountService.FindByID(testOnlyUserID)

	if err != nil {
		t.Fatalf("The testing user does not exist.")
	}

	{

		userCanSignIn := signInService.CanSignIn(user)
		assert.True(t, userCanSignIn, "Expected success to be true: CanSignIn")

		success, err := signInService.PasswordSignIn(user, testOnlyUserPw)
		assert.NoError(t, err)
		assert.True(t, success, "Expected success to be true: PasswordSignIn")

	}

	{
		type tcType struct {
			title    string
			user     service.UserAccount
			pw       string
			expected bool
		}
		items := []tcType{}

		{
			userTmp := *user
			items = append(items, tcType{
				title:    "test 0",
				user:     userTmp,
				pw:       testOnlyUserPw,
				expected: true, // check point
			})
		}

		{
			userTmp := *user
			userTmp.PasswordHash = ""
			items = append(items, tcType{
				title:    "test 1",
				user:     userTmp,
				pw:       testOnlyUserPw,
				expected: false,
			})
		}

		{
			userTmp := *user
			userTmp.PasswordHash = "123"
			items = append(items, tcType{
				title:    "test 2",
				user:     userTmp,
				pw:       testOnlyUserPw,
				expected: false,
			})
		}

		{
			userTmp := *user
			items = append(items, tcType{
				title:    "test 3",
				user:     userTmp,
				pw:       "",
				expected: false,
			})
		}

		{
			userTmp := *user
			items = append(items, tcType{
				title:    "test 4",
				user:     userTmp,
				pw:       testOnlyUserPw,
				expected: true, // check point
			})
		}

		for _, tc := range items {

			t.Run(tc.title, func(t *testing.T) {
				success, err := signInService.PasswordSignIn(&tc.user, tc.pw)
				assert.NoError(t, err)

				if success != tc.expected {
					t.Errorf("Expected %+v, got %+v", tc.expected, success)
				}
			})

		}

	}
}
