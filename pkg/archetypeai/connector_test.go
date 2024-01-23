package archetypeai

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util/httpclient"
	pb "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
	"github.com/instill-ai/x/errmsg"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	apiKey = "213bac"
)

const errJSON = `{ "error": "Invalid access." }`
const summarizeJSON = `
{
  "query_id": "240123b93a83a79e9907a5",
  "status": "completed",
  "file_ids": [
    "test_image.jpg"
  ],
  "inference_time_sec": 2.1776912212371826,
  "query_response_time_sec": 2.1914472579956055,
  "response": {
    "processed_text": "A family of four is hiking together on a trail."
  }
}`
const summarizeErrJSON = `
{
  "query_id": "2401233472bde249e60260",
  "status": "failed",
  "file_ids": [
    "test_image.jp"
  ]
}`

var (
	summarizeIn = summarizeParams{
		Query:   "Describe the image",
		FileIDs: []string{"test_image.jpg"},
	}
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name string

		task    string
		in      any
		want    any
		wantErr string

		// server expectations and response
		wantPath  string
		wantReq   any
		gotStatus int
		gotResp   string
	}{
		{
			name: "ok - summarize",

			task: taskSummarize,
			in:   summarizeIn,
			want: summarizeOutput{
				Response: "A family of four is hiking together on a trail.",
			},

			wantPath:  summarizePath,
			wantReq:   summarizeReq(summarizeIn),
			gotStatus: http.StatusOK,
			gotResp:   summarizeJSON,
		},
		{
			name: "nok - summarize wrong file",

			task:    taskSummarize,
			in:      summarizeIn,
			wantErr: `Archetype AI didn't complete query 2401233472bde249e60260: status is "failed".`,

			wantPath:  summarizePath,
			wantReq:   summarizeReq(summarizeIn),
			gotStatus: http.StatusOK,
			gotResp:   summarizeErrJSON,
		},
		{
			name: "nok - unauthorized",

			task:    taskSummarize,
			in:      summarizeIn,
			wantErr: "Archetype AI responded with a 401 status code. Invalid access.",

			wantPath:  summarizePath,
			wantReq:   summarizeReq(summarizeIn),
			gotStatus: http.StatusUnauthorized,
			gotResp:   errJSON,
		},
	}

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Matches, tc.wantPath)

				c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				w.WriteHeader(tc.gotStatus)
				fmt.Fprintln(w, tc.gotResp)
			})

			srv := httptest.NewServer(h)
			c.Cleanup(srv.Close)

			config, _ := structpb.NewStruct(map[string]any{
				"base_path": srv.URL,
				"api_key":   apiKey,
			})

			exec, err := connector.CreateExecution(defID, tc.task, config, logger)
			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.in)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execute([]*structpb.Struct{pbIn})
			if tc.wantErr != "" {
				c.Check(errmsg.Message(err), qt.Equals, tc.wantErr)
				return
			}

			c.Check(err, qt.IsNil)
			c.Assert(got, qt.HasLen, 1)

			wantJSON, err := json.Marshal(tc.want)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}
}

func TestConnector_CreateExecution(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	c.Run("nok - unsupported task", func(c *qt.C) {
		task := "FOOBAR"
		want := fmt.Sprintf("%s task is not supported.", task)

		_, err := connector.CreateExecution(defID, task, new(structpb.Struct), logger)
		c.Check(err, qt.IsNotNil)
		c.Check(errmsg.Message(err), qt.Equals, want)
	})
}

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	c.Run("ok - connected", func(c *qt.C) {
		got, err := connector.Test(defID, nil, logger)
		c.Check(err, qt.IsNil)
		c.Check(got, qt.Equals, pb.Connector_STATE_CONNECTED)
	})
}
