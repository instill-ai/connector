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
)

const (
	pineconeKey = "secret-key"

	upsertPath = "/vectors/upsert"
	upsertResp = `{"upsertedCount": 1}`

	queryPath = "/query"
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
			name: "ok - query",

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
	}

	logger := zap.NewNop()
	connector := Init(logger)

	for _, tc := range testcases {
		c.Run(tc.name, func(c *qt.C) {
			pineconeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// For now only POST methods are considered. When this changes,
				// this will need to be asserted per-path.
				c.Check(r.Method, qt.Equals, "POST")
				c.Check(r.URL.Path, qt.Equals, tc.wantClientPath)

				c.Check(r.Header.Get("Content-Type"), qt.Equals, jsonMimeType)
				c.Check(r.Header.Get("Accept"), qt.Equals, jsonMimeType)
				c.Check(r.Header.Get("Api-Key"), qt.Equals, pineconeKey)

				c.Assert(r.Body, qt.IsNotNil)
				defer r.Body.Close()

				body, err := io.ReadAll(r.Body)
				c.Assert(err, qt.IsNil)
				c.Check(body, qt.JSONEquals, tc.wantClientReq)

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)

				fmt.Fprintln(w, tc.clientResp)
			}))
			c.Cleanup(pineconeServer.Close)

			config, err := structpb.NewStruct(map[string]any{
				"api_key": pineconeKey,
				"url":     pineconeServer.URL,
			})

			defID := uuid.Must(uuid.NewV4())
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
}
