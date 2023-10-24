package instill

import (
	"crypto/tls"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

// initModelPublicServiceClient initialises a ModelPublicServiceClient instance
func initModelPublicServiceClient(serverURL string) (modelPB.ModelPublicServiceClient, *grpc.ClientConn) {
	var clientDialOpts grpc.DialOption

	if strings.HasPrefix(serverURL, "https://") {
		clientDialOpts = grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true}))
	} else {
		clientDialOpts = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	serverURL = stripProtocolFromURL(serverURL)
	clientConn, err := grpc.Dial(serverURL, clientDialOpts)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	return modelPB.NewModelPublicServiceClient(clientConn), clientConn
}

func stripProtocolFromURL(url string) string {
	index := strings.Index(url, "://")
	if index > 0 {
		return url[strings.Index(url, "://")+3:]
	}
	return url
}
