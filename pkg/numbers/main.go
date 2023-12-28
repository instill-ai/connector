package numbers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	_ "embed"
	b64 "encoding/base64"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/instill-ai/component/pkg/base"

	pipelinePB "github.com/instill-ai/protogen-go/vdp/pipeline/v1beta"
)

const ApiUrlPin = "https://api.numbersprotocol.io/api/v3/assets/"
const ApiUrlCommit = "https://eo883tj75azolos.m.pipedream.net"
const ApiUrlMe = "https://api.numbersprotocol.io/api/v3/auth/users/me"

var once sync.Once
var connector base.IConnector

//go:embed config/definitions.json
var definitionsJSON []byte

//go:embed config/tasks.json
var tasksJSON []byte

type Connector struct {
	base.Connector
}

type Execution struct {
	base.Execution
}

type CommitCustomLicense struct {
	Name     *string `json:"name,omitempty"`
	Document *string `json:"document,omitempty"`
}
type CommitCustom struct {
	DigitalSourceType *string              `json:"digitalSourceType,omitempty"`
	MiningPreference  *string              `json:"miningPreference,omitempty"`
	GeneratedThrough  string               `json:"generatedThrough"`
	GeneratedBy       *string              `json:"generatedBy,omitempty"`
	CreatorWallet     *string              `json:"creatorWallet,omitempty"`
	License           *CommitCustomLicense `json:"license,omitempty"`
	Metadata          *struct {
		Pipeline *struct {
			Uid    *string     `json:"uid,omitempty"`
			Recipe interface{} `json:"recipe,omitempty"`
		} `json:"pipeline,omitempty"`
		Owner *struct {
			Uid *string `json:"uid,omitempty"`
		} `json:"owner,omitempty"`
	} `json:"instillMetadata,omitempty"`
}
type Commit struct {
	AssetCid              string        `json:"assetCid"`
	AssetSha256           string        `json:"assetSha256"`
	EncodingFormat        string        `json:"encodingFormat"`
	AssetTimestampCreated int64         `json:"assetTimestampCreated"`
	AssetCreator          *string       `json:"assetCreator,omitempty"`
	Abstract              *string       `json:"abstract,omitempty"`
	Headline              *string       `json:"headline,omitempty"`
	Custom                *CommitCustom `json:"custom,omitempty"`
	Testnet               bool          `json:"testnet"`
}

type Input struct {
	Images       []string `json:"images"`
	AssetCreator *string  `json:"asset_creator,omitempty"`
	Abstract     *string  `json:"abstract,omitempty"`
	Headline     *string  `json:"headline,omitempty"`
	Custom       *struct {
		DigitalSourceType *string `json:"digital_source_type,omitempty"`
		MiningPreference  *string `json:"mining_preference,omitempty"`
		GeneratedBy       *string `json:"generated_by,omitempty"`
		License           *struct {
			Name     *string `json:"name,omitempty"`
			Document *string `json:"document,omitempty"`
		} `json:"license,omitempty"`
		Metadata *struct {
			Pipeline *struct {
				Uid    *string     `json:"uid,omitempty"`
				Recipe interface{} `json:"recipe,omitempty"`
			} `json:"pipeline,omitempty"`
			Owner *struct {
				Uid *string `json:"uid,omitempty"`
			} `json:"owner,omitempty"`
		} `json:"metadata,omitempty"`
	} `json:"custom,omitempty"`
}

type Output struct {
	AssetUrls []string `json:"asset_urls"`
}

func Init(logger *zap.Logger) base.IConnector {
	once.Do(func() {

		connector = &Connector{
			Connector: base.Connector{
				Component: base.Component{Logger: logger},
			},
		}
		err := connector.LoadConnectorDefinitions(definitionsJSON, tasksJSON, nil)
		if err != nil {
			logger.Fatal(err.Error())
		}

	})
	return connector
}

func getToken(config *structpb.Struct) string {
	return fmt.Sprintf("token %s", config.GetFields()["capture_token"].GetStringValue())
}

func (con *Execution) pinFile(data []byte) (string, string, error) {

	var b bytes.Buffer

	w := multipart.NewWriter(&b)
	var fw io.Writer
	var err error

	fileName, _ := uuid.NewV4()
	if fw, err = w.CreateFormFile("asset_file", fileName.String()+mimetype.Detect(data).Extension()); err != nil {
		return "", "", err
	}

	if _, err := io.Copy(fw, bytes.NewReader(data)); err != nil {
		return "", "", err
	}

	h := sha256.New()

	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return "", "", err
	}

	w.Close()
	sha256hash := fmt.Sprintf("%x", h.Sum(nil))

	req, err := http.NewRequest("POST", ApiUrlPin, &b)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", getToken(con.Config))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return "", "", err
	}

	if res.StatusCode == http.StatusCreated {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", "", err
		}
		var jsonRes map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &jsonRes)
		if cid, ok := jsonRes["cid"]; ok {
			return cid.(string), sha256hash, nil
		} else {
			return "", "", fmt.Errorf("pinFile failed")
		}

	} else {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", "", err
		}
		return "", "", fmt.Errorf(string(bodyBytes))
	}
}

func (con *Execution) commit(commit Commit) (string, string, error) {

	marshalled, err := json.Marshal(commit)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest("POST", ApiUrlCommit, bytes.NewReader(marshalled))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", getToken(con.Config))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return "", "", err
	}

	if res.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", "", err
		}
		var jsonRes map[string]interface{}
		_ = json.Unmarshal(bodyBytes, &jsonRes)

		var assetCid string
		var assetTreeCid string
		if val, ok := jsonRes["assetCid"]; ok {
			assetCid = val.(string)
		} else {
			return "", "", fmt.Errorf("assetCid failed")
		}
		if val, ok := jsonRes["assetTreeCid"]; ok {
			assetTreeCid = val.(string)
		} else {
			return "", "", fmt.Errorf("assetTreeCid failed")
		}
		return assetCid, assetTreeCid, nil

	} else {
		bodyBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return "", "", err
		}
		return "", "", fmt.Errorf(string(bodyBytes))
	}

}
func (c *Connector) CreateExecution(defUID uuid.UUID, task string, config *structpb.Struct, logger *zap.Logger) (base.IExecution, error) {
	e := &Execution{}
	e.Execution = base.CreateExecutionHelper(e, c, defUID, task, config, logger)
	return e, nil
}

func (e *Execution) Execute(inputs []*structpb.Struct) ([]*structpb.Struct, error) {

	var outputs []*structpb.Struct

	for _, input := range inputs {

		assetUrls := []string{}

		inputStruct := Input{}
		err := base.ConvertFromStructpb(input, &inputStruct)
		if err != nil {
			return nil, err
		}

		for _, image := range inputStruct.Images {
			imageBytes, err := b64.StdEncoding.DecodeString(base.TrimBase64Mime(image))
			if err != nil {
				return nil, err
			}

			var commitCustom *CommitCustom
			if inputStruct.Custom != nil {
				var commitCustomLicense *CommitCustomLicense
				if inputStruct.Custom.License != nil {
					commitCustomLicense = &CommitCustomLicense{
						Name:     inputStruct.Custom.License.Name,
						Document: inputStruct.Custom.License.Document,
					}
				}
				commitCustom = &CommitCustom{
					DigitalSourceType: inputStruct.Custom.DigitalSourceType,
					MiningPreference:  inputStruct.Custom.MiningPreference,
					GeneratedThrough:  "https://console.instill.tech", //TODO: support Core Host
					GeneratedBy:       inputStruct.Custom.GeneratedBy,
					License:           commitCustomLicense,
					Metadata:          inputStruct.Custom.Metadata,
				}

			}

			cid, sha256hash, err := e.pinFile(imageBytes)
			if err != nil {
				return nil, err
			}

			assetCid, _, err := e.commit(Commit{
				AssetCid:              cid,
				AssetSha256:           sha256hash,
				EncodingFormat:        http.DetectContentType(imageBytes),
				AssetTimestampCreated: time.Now().Unix(),
				AssetCreator:          inputStruct.AssetCreator,
				Abstract:              inputStruct.Abstract,
				Headline:              inputStruct.Headline,
				Custom:                commitCustom,
				Testnet:               false,
			})

			if err != nil {
				return nil, err
			}
			assetUrls = append(assetUrls, fmt.Sprintf("https://verify.numbersprotocol.io/asset-profile?nid=%s", assetCid))
		}

		outputStruct := Output{
			AssetUrls: assetUrls,
		}

		output, err := base.ConvertToStructpb(outputStruct)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)

	}

	return outputs, nil

}

func (con *Connector) Test(defUid uuid.UUID, config *structpb.Struct, logger *zap.Logger) (pipelinePB.Connector_State, error) {

	req, err := http.NewRequest("GET", ApiUrlMe, nil)
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, nil
	}
	req.Header.Set("Authorization", getToken(config))

	tr := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: tr}
	res, err := client.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return pipelinePB.Connector_STATE_ERROR, nil
	}
	if res.StatusCode == http.StatusOK {
		return pipelinePB.Connector_STATE_CONNECTED, nil
	}
	return pipelinePB.Connector_STATE_ERROR, nil
}
