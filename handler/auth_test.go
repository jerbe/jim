package handler

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/9/1 17:07
  @describe :
*/

func TestAuthLoginHandler(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AuthLoginHandler(tt.args.c)
		})
	}
}

func TestAuthLogoutHandler(t *testing.T) {
	type args struct {
		c *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AuthLogoutHandler(tt.args.c)
		})
	}
}

func TestAuthRegisterHandler(t *testing.T) {
	r := gin.New()
	jsonData := `{
		"username":"%s",
		"password":"%s",
		"confirm_password":"%s",
		"nickname":"%s",
		"birth_date":"%s"
	}`
	var jsonChan = make(chan string)
	var now = time.Now()

	go func() {
		for i := 0; i < 100; i++ {
			un := fmt.Sprintf("user_%06x", i)
			jsonChan <- fmt.Sprintf(jsonData, un, "password", "password", un, now.Format("2006-01-02"))
			now = now.AddDate(0, 0, 1)

			if i == 100-1 {
				jsonChan <- "end"
			}
		}
	}()

	for s := range jsonChan {
		if s == "end" {
			break
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/auth/register", bytes.NewBufferString(s))
		req.Header.Add("Content-Type", "application/json")

		ctx := gin.CreateTestContextOnly(w, r)
		ctx.Request = req

		RequestLogMiddleware()(ctx)
		AuthRegisterHandler(ctx)

		defer w.Result().Body.Close()

		io.Copy(os.Stdout, w.Result().Body)
	}

}
