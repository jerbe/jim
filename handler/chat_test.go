package handler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/gorilla/websocket"
	goutils "github.com/jerbe/go-utils"
	"sync"
	"time"

	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

/**
  @author : Jerbe - The porter from Earth
  @time : 2023/8/16 19:15
  @describe :
*/

func TestChatDeleteMessage(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DeleteChatMessageHandler(tt.args.ctx)
		})
	}
}

func TestChatRollbackMessage(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RollbackChatMessageHandler(tt.args.ctx)
		})
	}
}

func TestChatSendMessage(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SendChatMessageHandler(tt.args.ctx)
		})
	}
}

func getToken() (string, error) {
	url := "http://127.0.0.1:8080/auth/login"
	method := "POST"

	payload := strings.NewReader(`{
    "username":"root",
    "password":"root"
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	m := make(map[string]interface{})
	err = json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		return "", err
	}
	return (m["data"]).(map[string]any)["token"].(string), nil
}

func BenchmarkChatSendMessageUseApi(b *testing.B) {
	//r := gin.New()
	jsonData := `{
    "action_id":"随机字符串",
    "target_id":2,
    "session_type":1,
    "type":1,
    "body":{"text":"test"}
}`

	//token, err := getToken()
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MjU2MzQ0ODl9.WTqdIHkA8D8lfE_nJVir9Z64Cy1gZ-V11extOlvUjSI"

	file, err := os.Open("./benchmark.log")
	if err != nil && !os.IsNotExist(err) {
		log.Println(err)
		return
	}
	os.Remove("./benchmark.log")

	file, err = os.OpenFile("./benchmark.log", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	httpCli := http.DefaultClient

	b.SetParallelism(10000)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			//w := httptest.NewRecorder()
			///req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/chat/send_message", bytes.NewBufferString(jsonData))
			b.StopTimer()
			req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/v1/chat/message/send", bytes.NewBufferString(jsonData))
			req.Header.Add("Authorization", token)
			req.Header.Add("Content-Type", "application/json")

			//ctx := gin.CreateTestContextOnly(w, r)
			//ctx.Request = req

			b.StartTimer()
			rsp, err := httpCli.Do(req)
			if err != nil {
				file.WriteString(err.Error() + "\n")
				continue
			}

			//CheckAuthMiddleware()(ctx)
			//SendChatMessageHandler(ctx)

			//rs := make(map[string]any)
			go func(writer io.Writer, reader io.ReadCloser) {
				defer reader.Close()
				_, err = io.Copy(writer, reader)
				if err != nil {
					log.Println("129", err)

				}
				_, err = writer.Write([]byte("\n"))
				if err != nil {
					log.Println("134", err)
				}
			}(file, rsp.Body)

			//err := json.NewDecoder(w.Body).Decode(&rs)
			//if err != nil {
			//
			//	io.
			//		log.Println(err)
			//}
			//fmt.Println(rs)

		}
	})

}

func BenchmarkChatSendMessage(b *testing.B) {
	r := gin.New()
	jsonData := `{
    "action_id":"随机字符串",
    "target_id":"2",
    "session_type":1,
    "type":1,
    "body":{"text":"test"}
}`

	token, err := getToken()
	if err != nil {
		log.Println(err)
		return
	}

	file, err := os.Open("./benchmark.log")
	if err != nil && os.IsNotExist(err) {
		log.Println(err)
		return
	}
	os.Remove("./benchmark.log")

	file, err = os.OpenFile("./benchmark.log", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "http://127.0.0.1:8080/api/chat/send_message", bytes.NewBufferString(jsonData))
		req.Header.Add("Authorization", token)
		req.Header.Add("Content-Type", "application/json")

		ctx := gin.CreateTestContextOnly(w, r)
		ctx.Request = req

		b.StartTimer()

		CheckAuthMiddleware()(ctx)
		SendChatMessageHandler(ctx)

		go func(writer io.Writer, reader io.ReadCloser) {
			defer reader.Close()
			_, err = io.Copy(writer, reader)
			if err != nil {
				log.Println("129", err)
			}
			_, err = writer.Write([]byte("\n"))
			if err != nil {
				log.Println("134", err)
			}
		}(file, w.Result().Body)

	}
}

func BenchmarkChatSendMessageParallel(b *testing.B) {
	r := gin.New()
	jsonData := `{
    "action_id":"随机字符串",
    "target_id":2,
    "session_type":1,
    "type":1,
    "body":{"text":"test"}
}`

	//token, err := getToken()
	//if err != nil {
	//	log.Println(err)
	//	return
	//}

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MjU2MzQ0ODl9.WTqdIHkA8D8lfE_nJVir9Z64Cy1gZ-V11extOlvUjSI"

	//file, err := os.Open("./benchmark.log")
	//if err != nil && os.IsNotExist(err) {
	//	log.Println(err)
	//	return
	//}

	limiterMiddleware := RateLimitMiddleware(goutils.NewLimiter(1<<2, time.Millisecond*100))
	os.Remove("./benchmark.log")

	file, err := os.OpenFile("./benchmark.log", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println(err)
		return
	}
	defer file.Close()

	b.SetParallelism(10000)
	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {

			w := httptest.NewRecorder()
			b.StopTimer()
			req := httptest.NewRequest(http.MethodPost, "http://192.168.31.100:8080/api/chat/send_message", bytes.NewBufferString(jsonData))
			req.Header.Add("Authorization", token)
			req.Header.Add("Content-Type", "application/json")

			ctx := gin.CreateTestContextOnly(w, r)
			ctx.Request = req

			b.StartTimer()
			limiterMiddleware(ctx)
			CheckAuthMiddleware()(ctx)
			SendChatMessageHandler(ctx)

			b.StopTimer()
			go func(writer io.Writer, reader io.ReadCloser) {
				defer reader.Close()
				_, err = io.Copy(writer, reader)
				if err != nil {
					log.Println("129", err)

				}
				_, err = writer.Write([]byte("\n"))
				if err != nil {
					log.Println("134", err)
				}
			}(file, w.Result().Body)
			b.StartTimer()
		}
	})

}

func TestWebsocketHandler(t *testing.T) {
	type args struct {
		ctx *gin.Context
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WebsocketHandler(tt.args.ctx)
		})
	}
}

func TestBenchmarkWebsocketApi(t *testing.T) {
	//token, err := getToken()
	//if err != nil {
	//	t.Fatal(err)
	//}

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJleHAiOjE3MjU2MzQ0ODl9.WTqdIHkA8D8lfE_nJVir9Z64Cy1gZ-V11extOlvUjSI"
	wg := &sync.WaitGroup{}

	for i := 0; i < 9999; i++ {
		wg.Add(1)
		go func() {
			defer func() {
				wg.Done()
				log.Println("协程已关闭")
			}()
			header := http.Header{}
			header.Add("Authorization", token)
			conn, rsp, err := websocket.DefaultDialer.Dial("ws://192.168.31.100:8080/api/ws", header)
			if err != nil {
				log.Println(err)
				return
			}
			defer func() {
				if conn != nil {
					conn.Close()
				}
				if rsp != nil {
					rsp.Body.Close()
				}
			}()
			if rsp.StatusCode != http.StatusSwitchingProtocols {
				log.Println("链接错误,不是正确的状态码: 需要101 但是得到:", rsp.StatusCode)
				return
			}

			for {
				typ, msg, err := conn.ReadMessage()
				if err != nil {
					log.Println("获取消息失败", err)
					return
				}
				log.Printf("类型:%d ,内容:%s", typ, msg)
			}
		}()
	}
	wg.Wait()
}
