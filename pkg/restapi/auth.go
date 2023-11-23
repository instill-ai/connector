package restapi

import (
	"encoding/base64"
	"errors"
)

type AuthType string

const (
	NoAuthType      AuthType = "NO_AUTH"
	BasicAuthType   AuthType = "BASIC_AUTH"
	APIKeyType      AuthType = "API_KEY"
	BearerTokenType AuthType = "BEARER_TOKEN"
)

// Authentication authentication interface
type Authentication interface {
	// GetAuthLocation returns the location of the authentication string, it can be "header" or "query"
	GetAuthLocation() AuthLocation
	// GenAuthHeader generates the authentication header key and value
	GenAuthHeader() (key, value string, err error)
	// GenAuthQuery generates the authentication key and value for query parameter
	GenAuthQuery() (key, value string, err error)
}

// NoAuth is the no authentication method
type NoAuth struct {
	AuthType AuthType `json:"auth_type"`
}

func (na NoAuth) GetAuthLocation() AuthLocation {
	return Header
}

func (na NoAuth) GenAuthHeader() (string, string, error) {
	return "", "", nil
}

func (na NoAuth) GenAuthQuery() (string, string, error) {
	return "", "", nil
}

// BasicAuth is the basic authentication method
type BasicAuth struct {
	AuthType AuthType `json:"auth_type"`
	Username string   `json:"username"`
	Password string   `json:"password"`
}

func (ba BasicAuth) GetAuthLocation() AuthLocation {
	return Header
}

func (ba BasicAuth) GenAuthHeader() (string, string, error) {
	if ba.Username == "" || ba.Password == "" {
		return "", "", errors.New("Basic Auth error: username or password is empty")
	}
	key := "Authorization"
	value := "Basic " + base64.StdEncoding.EncodeToString([]byte(ba.Username+":"+ba.Password))
	return key, value, nil
}

func (ba BasicAuth) GenAuthQuery() (string, string, error) {
	return "", "", errors.New("Basic Auth error: Basic Auth does not support query parameter")
}

// AuthLocation is the enum for AuthLocation field in ApiKeyAuth struct, which represents the location of the authentication string, it can be "header" or "query"
type AuthLocation string

const (
	Header AuthLocation = "header"
	Query  AuthLocation = "query"
)

// ApiKeyAuth is the API key authentication method
type APIKeyAuth struct {
	AuthType     AuthType     `json:"auth_type"`
	Key          string       `json:"key"`
	Value        string       `json:"value"`
	AuthLocation AuthLocation `json:"auth_location"`
}

func (aka APIKeyAuth) GetAuthLocation() AuthLocation {
	return aka.AuthLocation
}

func (aka APIKeyAuth) GenAuthHeader() (string, string, error) {
	if aka.Key == "" || aka.Value == "" {
		return "", "", errors.New("API Key Auth error: key or value is empty")
	}
	return aka.Key, aka.Value, nil
}

func (aka APIKeyAuth) GenAuthQuery() (string, string, error) {
	if aka.Key == "" || aka.Value == "" {
		return "", "", errors.New("API Key Auth error: key or value is empty")
	}
	return aka.Key, aka.Value, nil
}

// BearerTokenAuth is the bearer token authentication method
type BearerTokenAuth struct {
	AuthType AuthType `json:"auth_type"`
	Token    string   `json:"token"`
}

func (bta BearerTokenAuth) GetAuthLocation() AuthLocation {
	return Header
}

func (bta BearerTokenAuth) GenAuthHeader() (string, string, error) {
	if bta.Token == "" {
		return "", "", errors.New("Bearer Token Auth error: token is empty")
	}
	key := "Authorization"
	value := "Bearer " + bta.Token
	return key, value, nil
}

func (bta BearerTokenAuth) GenAuthQuery() (string, string, error) {
	return "", "", errors.New("Bearer Token Auth error: Bearer Token Auth does not support query parameter")
}
