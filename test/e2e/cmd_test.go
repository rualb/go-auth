package e2e

import (
	"encoding/json"
	"fmt"
	xcmd "go-auth/internal/cmd"
	"go-auth/internal/config/consts"
	"go-auth/internal/service"
	"go-auth/internal/util/utilhttp"
	"log"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testOnlyUserID = "test-only-1"
const testOnlyUserTel = "+999999999999"
const testOnlyUserPw = "Qq123456"

var appService service.AppService

func setup() func() {

	os.Setenv("APP_ENV", "testing")
	os.Setenv("APP_IDENTITY_IS_AUTH_TEL", "true")
	os.Setenv("APP_IDENTITY_TEL_PREFIX", "+999")

	cmd := xcmd.Command{}

	go cmd.Exec()

	time.Sleep(1 * time.Second)

	appService = cmd.AppService

	db := appService.Repository()

	{
		//
		testUserAcc, _ := service.NewUserAccount()
		testUserAcc.ID = testOnlyUserID
		testUserAcc.Tel = testOnlyUserTel
		testUserAcc.SetPassword(testOnlyUserPw)
		//
		res := db.Save(&testUserAcc)
		if res.Error != nil {
			log.Fatalf("cannot create test-only user")
		}
		//
	}

	return cmd.Stop
}

func TestMain(m *testing.M) {
	fmt.Println("starting test suite...")
	stop := setup()
	defer func() {
		fmt.Println("cleaning up test environment...")
		stop()
	}()

	fmt.Println("running tests...")
	exitCode := m.Run()

	fmt.Println("test suite complete!")
	os.Exit(exitCode)
}

func getAuthToken(srv service.AppService, iss string, expire bool) (authToken string) {

	var timOffset time.Duration = 0

	if expire {
		timOffset = time.Hour * 24 * 365 * -1
	}

	secret, err := srv.Vault().CurrentKey()
	if err != nil {
		log.Fatalf("error getting secret: %v", err)
	}
	token := jwt.New(jwt.SigningMethodHS256)
	token.Header["kid"] = secret.ID
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = testOnlyUserID
	claims["iat"] = time.Now().Add(timOffset).Unix()
	claims["exp"] = time.Now().Add(time.Hour * 720).Add(timOffset).Unix()
	claims["iss"] = iss

	tokenString, err := token.SignedString(secret.AuthKey)
	if err != nil {
		log.Fatalf("error signing token: %v", err)
	}

	return tokenString

}
func TestAuthTokenDump(t *testing.T) {

	// // /app/home/postgres/bin/psql -d postgres -U postgres -c "update user_accounts set roles='admin' where id = 'test-only-1'"

	tkn := getAuthToken(appService, "auth", false)

	t.Log(tkn)
}
func TestAuthHTTP(t *testing.T) {

	urlAPI := "http://127.0.0.1:30280" + consts.PathAuthStatusAPI

	header := map[string]string{
		"Cookie": "_auth=" + url.QueryEscape(getAuthToken(appService, "auth", false)),
	}
	headerNotAuth := map[string]string{
		"Cookie": "_auth=" + url.QueryEscape(getAuthToken(appService, "not-auth", false)),
	}
	headerExpired := map[string]string{
		"Cookie": "_auth=" + url.QueryEscape(getAuthToken(appService, "auth", true)),
	}

	type StatusResponse struct {
		IsAuth bool `json:"is_auth"`
	}

	urls := []struct {
		title    string
		url      string
		query    map[string]string
		header   map[string]string
		expected StatusResponse
	}{
		{title: "test_auth", expected: StatusResponse{IsAuth: true},
			url:   urlAPI,
			query: map[string]string{}, header: header},
		{title: "test_no_auth", expected: StatusResponse{IsAuth: false},
			url:   urlAPI,
			query: map[string]string{}, header: headerNotAuth},
		{title: "test_no_header", expected: StatusResponse{IsAuth: false},
			url:   urlAPI,
			query: map[string]string{}, header: nil},
		{title: "test_expired", expected: StatusResponse{IsAuth: false},
			url:   urlAPI,
			query: map[string]string{}, header: headerExpired},
	}

	for _, itm := range urls {

		t.Run(itm.title, func(t *testing.T) {

			t.Logf("url %v", itm.url)
			dataArr, err := utilhttp.GetBytes(itm.url, itm.query, itm.header)
			if err != nil {
				t.Errorf("Error : %v", err)
			}

			var response StatusResponse
			if err := json.Unmarshal(dataArr, &response); err != nil {
				t.Errorf("Failed to unmarshal response: %v", err)
			}

			if !reflect.DeepEqual(response, itm.expected) {
				t.Errorf("Expected %+v, got %+v", itm.expected, response)
			}

		})

	}

}
