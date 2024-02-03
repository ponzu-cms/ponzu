package entities

type CredentialType string

var CredentialTypePassword CredentialType = "password"

type PasswordHash struct {
	Hash string `json:"hash"`
	Salt string `json:"salt"`
}

type Credential struct {
	Type CredentialType `json:"type"`
	// TODO: value should be an interface. supporting many credential value types
	Value string `json:"value"`
}

type CredentialHash struct {
	UserId string         `json:"user_id"`
	Type   CredentialType `json:"type"`
	Value  []byte         `json:"value"`
}
