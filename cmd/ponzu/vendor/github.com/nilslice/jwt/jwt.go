package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	privateKey       = []byte("")
	defaultHeader    = header{Typ: "JWT", Alg: "HS256"}
	registeredClaims = []string{"iss", "sub", "aud", "exp", "nbf", "iat", "jti"}
)

type header struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

type payload map[string]interface{}

type encoded struct {
	token string
}
type decoded struct {
	header  string
	payload string
}

type signedDecoded struct {
	decoded
	signature string
}

func newEncoded(claims map[string]interface{}) (encoded, error) {
	header, err := json.Marshal(defaultHeader)
	if err != nil {
		return encoded{}, err
	}

	for _, claim := range registeredClaims {
		if _, ok := claims[claim]; !ok {
			claims[claim] = nil
		}
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return encoded{}, err
	}

	d := decoded{header: string(header), payload: string(payload)}

	d.encodeInternal()
	signed, err := d.sign()
	if err != nil {
		return encoded{}, err
	}

	token := signed.token()
	e := encoded{token: token}
	return e, nil
}

func newDecoded(token string) (decoded, error) {
	e := encoded{token: token}
	d, err := e.parseToken()
	if err != nil {
		return d, nil
	}

	return d, nil
}

func encodeToString(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

func (d decoded) getHeader() []byte {
	return []byte(d.header)
}

func (d decoded) getPayload() []byte {
	return []byte(d.payload)
}

func (sd signedDecoded) getSignature() []byte {
	return []byte(sd.signature)
}

func (d *decoded) encodeInternal() {
	d.header = encodeToString(d.getHeader())
	d.payload = encodeToString(d.getPayload())
}

func (d decoded) dot(internals ...string) string {
	return strings.Join(internals, ".")
}

func (d *decoded) sign() (signedDecoded, error) {
	if d.header == "" || d.payload == "" {
		return signedDecoded{}, errors.New("Missing header or payload on Decoded")
	}

	unsigned := d.dot(d.header, d.payload)

	hash := hmac.New(sha256.New, privateKey)
	_, err := hash.Write([]byte(unsigned))
	if err != nil {
		return signedDecoded{}, err
	}

	signed := signedDecoded{decoded: *d}
	signed.signature = encodeToString(hash.Sum(nil))

	return signed, nil
}

func (sd signedDecoded) token() string {
	return fmt.Sprintf("%s.%s.%s", sd.getHeader(), sd.getPayload(), sd.getSignature())
}

func (sd signedDecoded) verify(enc encoded) bool {
	if sd.token() == enc.token {
		return true
	}
	return false
}

func (e encoded) parseToken() (decoded, error) {
	parts := strings.Split(e.token, ".")
	if len(parts) != 3 {
		return decoded{}, errors.New("Error: incorrect # of results from string parsing")
	}

	d := decoded{
		header:  parts[0],
		payload: parts[1],
	}

	return d, nil
}

// New returns a token (string) and error. The token is a fully qualified JWT to be sent to a client via HTTP Header or other method. Error returned will be from the newEncoded unexported function.
func New(claims map[string]interface{}) (string, error) {
	enc, err := newEncoded(claims)
	if err != nil {
		return "", err
	}

	return enc.token, nil
}

// Passes returns a bool indicating whether a token (string) provided has been signed by our server. If true, the client is authenticated and may proceed.
func Passes(token string) bool {
	dec, err := newDecoded(token)
	if err != nil {
		// may want to log some error here so we have visibility
		// intentionally simplifying return type to bool for ease
		// of use in API. Caller should only do `if auth.Passes(str) {}`
		return false
	}
	signed, err := dec.sign()
	if err != nil {
		return false
	}

	return signed.verify(encoded{token: token})
}

// GetClaims() returns a token's claims, allowing
// you to check the values to make sure they match
func GetClaims(token string) map[string]interface{} {
	// decode the token
	dec, err := newDecoded(token)
	if err != nil {
		return nil
	}

	// base64 decode payload
	payload, err := base64.RawURLEncoding.DecodeString(dec.payload)
	if err != nil {
		return nil
	}

	dst := map[string]interface{}{}
	err = json.Unmarshal(payload, &dst)
	if err != nil {
		return nil
	}

	return dst

}

// Secret is a helper function to set the unexported privateKey variable used when signing and verifying tokens.
// Its argument is type []byte since we expect users to read this value from a file which can be excluded from source code.
func Secret(key []byte) {
	privateKey = key
}
