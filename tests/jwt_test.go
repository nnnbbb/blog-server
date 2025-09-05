package tests

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"blog-server/server" // 替换为你的 go.mod module 名

	"blog-server/config"

	"github.com/stretchr/testify/assert"
)

func TestLogin_Success(t *testing.T) {

	config.Init("")

	r := server.NewRouter()

	form := url.Values{}
	form.Add("username", "admin")
	form.Add("password", "123456")

	req := httptest.NewRequest("POST", "/v1/user/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token")
}
