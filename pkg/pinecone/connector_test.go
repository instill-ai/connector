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
)

var (
	vectorA = Vector{
		ID:       "A",
		Values:   []float64{2.23},
		Metadata: map[string]any{"color": "pumpkin"},
	}
)

func TestConnector_Execute(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)

	pineconeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Check(r.URL.Path, qt.Equals, upsertPath)
		c.Check(r.Method, qt.Equals, "POST")

		c.Check(r.Header.Get("Content-Type"), qt.Equals, jsonMimeType)
		c.Check(r.Header.Get("Accept"), qt.Equals, jsonMimeType)
		c.Check(r.Header.Get("Api-Key"), qt.Equals, pineconeKey)

		c.Assert(r.Body, qt.IsNotNil)
		defer r.Body.Close()

		body, err := io.ReadAll(r.Body)
		c.Assert(err, qt.IsNil)
		c.Check(body, qt.JSONEquals, UpsertReq{Vectors: []Vector{vectorA}})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, upsertResp)
	}))
	c.Cleanup(pineconeServer.Close)

	config, err := structpb.NewStruct(map[string]any{
		"api_key": pineconeKey,
		"url":     pineconeServer.URL,
	})

	defID := uuid.Must(uuid.NewV4())
	exec, err := connector.CreateExecution(defID, taskUpsert, config, logger)
	c.Assert(err, qt.IsNil)

	in := vectorA
	pbIn, err := base.ConvertToStructpb(in)
	c.Assert(err, qt.IsNil)

	out, err := exec.Execute([]*structpb.Struct{pbIn})
	c.Check(err, qt.IsNil)

	c.Assert(out, qt.HasLen, 1)
	want, err := json.Marshal(UpsertOutput{
		RecordsUpserted: 1,
	})
	c.Assert(err, qt.IsNil)
	c.Check(want, qt.JSONEquals, out[0].AsMap())
}
