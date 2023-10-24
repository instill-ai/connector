package main

import (
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/gofrs/uuid"
	connector "github.com/instill-ai/connector/pkg"
)

func main() {

	logger, _ := zap.NewDevelopment()
	// It is singleton, should be loaded when connector-backend started
	connector := connector.Init(logger, connector.ConnectorOptions{})

	fmt.Println(connector.ListConnectorDefinitions())

	execution, _ := connector.CreateExecution(uuid.FromStringOrNil("9fb6a2cb-bff5-4c69-bc6d-4538dd8e3362"), "TASK_TEXT_GENERATION", &structpb.Struct{}, logger)

	r, err := execution.ExecuteWithValidation([]*structpb.Struct{})
	fmt.Println(r, err)

}
