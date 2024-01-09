package pinecone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/util/httpclient"
	"github.com/instill-ai/x/errmsg"
)

const (
	pineconeKey = "secret-key"

	upsertResp = `{"upsertedCount": 1}`

	queryResp = `
{
	"namespace": "color-schemes",
	"matches": [
		{
			"id": "A",
			"values": [ 2.23 ],
			"metadata": { "color": "pumpkin" },
			"score": 0.99
		}
	]
}`

	errResp = `
{
  "code": 3,
  "message": "Cannot provide both ID and vector at the same time.",
  "details": []
}`
)

var (
	vectorA = Vector{
		ID:       "A",
		Values:   []float64{2.23},
		Metadata: map[string]any{"color": "pumpkin"},
	}
	queryByVector = QueryInput{
		Namespace:       "color-schemes",
		TopK:            1,
		Vector:          vectorA.Values,
		IncludeValues:   true,
		IncludeMetadata: true,
		Filter: map[string]any{
			"color": map[string]any{
				"$in": []string{"green", "cerulean", "pumpkin"},
			},
		},
	}
	queryByID = QueryInput{
		Namespace:       "color-schemes",
		TopK:            1,
		Vector:          vectorA.Values,
		ID:              vectorA.ID,
		IncludeValues:   true,
		IncludeMetadata: true,
	}
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)

	testcases := []struct {
		name string

		task     string
		execIn   any
		wantExec any

		wantClientPath string
		wantClientReq  any
		clientResp     string
	}{
		{
			name: "ok - upsert",

			task:     taskUpsert,
			execIn:   vectorA,
			wantExec: UpsertOutput{RecordsUpserted: 1},

			wantClientPath: upsertPath,
			wantClientReq:  UpsertReq{Vectors: []Vector{vectorA}},
			clientResp:     upsertResp,
		},
		{
			name: "ok - query by vector",

			task:   taskQuery,
			execIn: queryByVector,
			wantExec: QueryResp{
				Namespace: "color-schemes",
				Matches: []Match{
					{
						Vector: vectorA,
						Score:  0.99,
					},
				},
			},

			wantClientPath: queryPath,
			wantClientReq:  QueryReq(queryByVector),
			clientResp:     queryResp,
		},
		{
			name: "ok - query by ID",

			task:   taskQuery,
			execIn: queryByID,
			wantExec: QueryResp{
				Namespace: "color-schemes",
				Matches: []Match{
					{
						Vector: vectorA,
						Score:  0.99,
					},
				},
			},

			wantClientPath: queryPath,
			wantClientReq: QueryReq{
				// Vector is wiped from the request.
				Namespace:       "color-schemes",
				TopK:            1,
				ID:              vectorA.ID,
				IncludeValues:   true,
				IncludeMetadata: true,
			},
			clientResp: queryResp,
		},
	}

	logger := zap.NewNop()
	connector := Init(logger)
	defID := uuid.Must(uuid.NewV4())

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// For now only POST methods are considered. When this changes,
				// this will need to be asserted per-path.
				c.Check(r.Method, qt.Equals, http.MethodPost)
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Accept"), qt.Equals, httpclient.MIMETypeJSON)
				c.Check(r.Header.Get("Api-Key"), qt.Equals, pineconeKey)

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", httpclient.MIMETypeJSON)
				fmt.Fprintln(w, tc.clientResp)
			})

			pineconeServer := httptest.NewServer(h)
			c.Cleanup(pineconeServer.Close)

			config, _ := structpb.NewStruct(map[string]any{
				"api_key": pineconeKey,
				"url":     pineconeServer.URL,
			})

			exec, err := connector.CreateExecution(defID, tc.task, config, logger)
			c.Assert(err, qt.IsNil)

			pbIn, err := base.ConvertToStructpb(tc.execIn)
			c.Assert(err, qt.IsNil)

			got, err := exec.Execute([]*structpb.Struct{pbIn})
			c.Check(err, qt.IsNil)

			c.Assert(got, qt.HasLen, 1)
			wantJSON, err := json.Marshal(tc.wantExec)
			c.Assert(err, qt.IsNil)
			c.Check(wantJSON, qt.JSONEquals, got[0].AsMap())
		})
	}

	c.Run("nok - 400", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintln(w, errResp)
		})

		pineconeServer := httptest.NewServer(h)
		c.Cleanup(pineconeServer.Close)

		config, _ := structpb.NewStruct(map[string]any{
			"url": pineconeServer.URL,
		})

		exec, err := connector.CreateExecution(defID, taskUpsert, config, logger)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "Pinecone responded with a 400 status code. Cannot provide both ID and vector at the same time."
		c.Check(errmsg.Message(err), qt.Equals, want)
	})

	c.Run("nok - URL misconfiguration", func(c *qt.C) {
		config, _ := structpb.NewStruct(map[string]any{
			"url": "http://no-such.host",
		})

		exec, err := connector.CreateExecution(defID, taskUpsert, config, logger)
		c.Assert(err, qt.IsNil)

		pbIn := new(structpb.Struct)
		_, err = exec.Execute([]*structpb.Struct{pbIn})
		c.Check(err, qt.IsNotNil)

		want := "Failed to call http://no-such.host/.*. Please check that the connector configuration is correct."
		c.Check(errmsg.Message(err), qt.Matches, want)
	})
}
