package restapi

type BodyType string

const (
	NoneBodyType BodyType = "NONE"
	RawJSONType  BodyType = "RAW_JSON"
	// TODO: add more body types
	// FormDataType BodyType = "FORM_DATA"
	// XWWWFORMURLENCODEDType BodyType = "X_WWW_FORM_URLENCODED"
	// BinaryType   BodyType = "BINARY"
)

type Body interface {
	GetBodyType() BodyType
	GetBody() map[string]interface{}
}

// NoneBody: body is none
type NoneBody struct {
	BodyType BodyType `json:"body_type"`
}

func (nb NoneBody) GetBodyType() BodyType {
	return NoneBodyType
}

func (nb NoneBody) GetBody() map[string]interface{} {
	return nil
}

// RawJSONBody: body is raw JSON
type RawJSONBody struct {
	BodyType BodyType               `json:"body_type"`
	BodyData map[string]interface{} `json:"data,omitempty"`
}

func (rjb RawJSONBody) GetBodyType() BodyType {
	return RawJSONType
}

func (rjb RawJSONBody) GetBody() map[string]interface{} {
	return rjb.BodyData
}
