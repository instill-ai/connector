package instill

import (
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
	"github.com/instill-ai/x/errmsg"
)

const (
	apiKey = "123"
)

func TestConnector_Test(t *testing.T) {
	c := qt.New(t)

	logger := zap.NewNop()
	connector := Init(logger)
	userID, defID := uuid.Must(uuid.NewV4()), uuid.Must(uuid.NewV4())

	wantPath := "/model/v1alpha/models"
	c.Run("nok - error", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, wantPath)

			c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
			c.Check(r.Header.Get("Instill-User-UID"), qt.Equals, userID.String())

			w.WriteHeader(http.StatusBadRequest)
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		config, err := structpb.NewStruct(map[string]any{
			"mode":             "external",
			"server_url":       srv.URL,
			"api_token":        apiKey,
			"instill_user_uid": userID.String(),
		})
		c.Assert(err, qt.IsNil)

		got, err := connector.Test(defID, config, logger)
		c.Check(err, qt.IsNotNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_ERROR)

		wantMsg := "Instill AI responded with a 400 status code. Please refer to Instill AI's API reference for more information."
		c.Check(errmsg.Message(err), qt.Equals, wantMsg)
	})

	c.Run("ok - connected", func(c *qt.C) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Check(r.Method, qt.Equals, http.MethodGet)
			c.Check(r.URL.Path, qt.Equals, wantPath)

			c.Check(r.Header.Get("Authorization"), qt.Equals, "Bearer "+apiKey)
			c.Check(r.Header.Get("Instill-User-UID"), qt.Equals, userID.String())
		})

		srv := httptest.NewServer(h)
		c.Cleanup(srv.Close)

		config, err := structpb.NewStruct(map[string]any{
			"mode":             "external",
			"server_url":       srv.URL,
			"api_token":        apiKey,
			"instill_user_uid": userID.String(),
		})
		c.Assert(err, qt.IsNil)

		got, err := connector.Test(defID, config, logger)
		c.Check(err, qt.IsNil)
		c.Check(got, qt.Equals, pipelinePB.Connector_STATE_CONNECTED)
	})
}
