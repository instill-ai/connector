package pkg

import (
	"sync"

	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"
	"github.com/instill-ai/connector/pkg/airbyte"
	"github.com/instill-ai/connector/pkg/bigquery"
	"github.com/instill-ai/connector/pkg/googlecloudstorage"
	"github.com/instill-ai/connector/pkg/googlesearch"
	"github.com/instill-ai/connector/pkg/huggingface"
	"github.com/instill-ai/connector/pkg/instill"
	"github.com/instill-ai/connector/pkg/numbers"
	"github.com/instill-ai/connector/pkg/openai"
	"github.com/instill-ai/connector/pkg/pinecone"
	"github.com/instill-ai/connector/pkg/redis"
	"github.com/instill-ai/connector/pkg/restapi"
	"github.com/instill-ai/connector/pkg/stabilityai"
	"github.com/instill-ai/connector/pkg/website"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

var once sync.Once
var connector base.IConnector

type Connector struct {
	base.Connector
	connectorUIDMap map[uuid.UUID]base.IConnector
}

type ConnectorOptions struct {
	Airbyte airbyte.ConnectorOptions
}

func Init(logger *zap.Logger, options ConnectorOptions) base.IConnector {
	once.Do(func() {

		connector = &Connector{
			Connector:       base.Connector{Component: base.Component{Logger: logger}},
			connectorUIDMap: map[uuid.UUID]base.IConnector{},
		}

		connector.(*Connector).ImportDefinitions(stabilityai.Init(logger))
		connector.(*Connector).ImportDefinitions(instill.Init(logger))
		connector.(*Connector).ImportDefinitions(huggingface.Init(logger))
		connector.(*Connector).ImportDefinitions(openai.Init(logger))
		connector.(*Connector).ImportDefinitions(numbers.Init(logger))
		connector.(*Connector).ImportDefinitions(airbyte.Init(logger, options.Airbyte))
		connector.(*Connector).ImportDefinitions(bigquery.Init(logger))
		connector.(*Connector).ImportDefinitions(googlecloudstorage.Init(logger))
		connector.(*Connector).ImportDefinitions(googlesearch.Init(logger))
		connector.(*Connector).ImportDefinitions(pinecone.Init(logger))
		connector.(*Connector).ImportDefinitions(redis.Init(logger))
		connector.(*Connector).ImportDefinitions(restapi.Init(logger))
		connector.(*Connector).ImportDefinitions(website.Init(logger))

	})
	return connector
}
func (c *Connector) ImportDefinitions(con base.IConnector) {
	for _, v := range con.ListConnectorDefinitions() {
		err := c.AddConnectorDefinition(v)
		if err != nil {
			panic(err)
		}
		c.connectorUIDMap[uuid.FromStringOrNil(v.Uid)] = con
	}
}

func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	return c.connectorUIDMap[defUID].CreateExecution(defUID, task, config, logger)
}

func (c *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {
	return c.connectorUIDMap[defUid].Test(defUid, config, logger)
}

func (c *Connector) GetOpenapiSpecificationsByUID(defUID uuid.UUID, config *structpb.Struct) *structpb.Struct {
	if uid := uuid.FromStringOrNil("5ee55a5c-6e30-4c7a-80e8-90165a729e0a"); defUID == uid {
		return c.connectorUIDMap[uid].(*restapi.Connector).GetOpenapiSpecificationsByUID(defUID, config)
	}
	return c.Connector.GetOpenapiSpecificationsByUID(defUID, config)
}
