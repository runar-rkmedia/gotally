package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/runar-rkmedia/go-common/logger"
	model "github.com/runar-rkmedia/gotally/gen/proto/tally/v1"
)

func TestApi_Restart(t *testing.T) {
	logger.InitLogger(logger.LogConfig{
		Level:      "debug",
		Format:     "human",
		WithCaller: true,
	})

	t.Run("Restart", func(t *testing.T) {
		// var connStr = fmt.Sprintf("file:%s?mode=memory&cache=shared", mustCreateUUidgenerator()())

		path, handler := createApiHandler(true, TallyOptions{
			// DatabaseDSN: "file::memory:?cache=shared",
			DatabaseDSN: "file:thisisauniquenamebutnotafile?mode=memory&cache=shared",
			// DatabaseDSN: connStr,
			// DatabaseDSN: "sqlite:./test.sqlite",
			SkipStatsCollection: true,
		})
		if false {

			time.Sleep(time.Second)
		}
		recorder := httptest.NewRecorder()
		// client := tallyv1connect.NewBoardServiceClient(http.DefaultClient, "http://localhost:2888")
		// client.RestartGame(context.TODO(), &connect.Request[model.RestartGameRequest]{})

		body := model.GetSessionRequest{}
		t.Error(mustJsonifyString(body))
		var p = path + "GetSession"
		req, err := http.NewRequest(http.MethodPost, p, mustJsonify(body))
		if err != nil {
			panic(err)
		}
		req.Header.Add("Accept", "application/json")
		req.Header.Add("Content-Type", "application/json")
		handler.ServeHTTP(recorder, req)
		resBody, err := io.ReadAll(recorder.Result().Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Log(p)
		t.Log(req.URL)
		t.Log(recorder.Result().Status)
		t.Log(recorder.Result().Header)
		t.Errorf("req %d, %s", recorder.Code, string(resBody))
	})
}

func mustJsonify(j interface{}) io.Reader {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	return bytes.NewReader(b)
}
func mustJsonifyString(j interface{}) string {
	b, err := json.Marshal(j)
	if err != nil {
		panic(err)
	}
	return string(b)
}
