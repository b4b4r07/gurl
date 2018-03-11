package iap

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	neturl "net/url"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jws"
)

const (
	// TokenURI is a token URL of Google API
	TokenURI = "https://www.googleapis.com/oauth2/v4/token"

	// GoogleApplicationCredentials is a credentials path of the service account
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
	// ClientID is client ID of the backend for IAP
	ClientID = "CLIENT_ID"
)

func readRsaPrivateKey(bytes []byte) (key *rsa.PrivateKey, err error) {
	block, _ := pem.Decode(bytes)
	if block == nil {
		err = errors.New("invalid private key data")
		return
	}

	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return
		}
	} else if block.Type == "PRIVATE KEY" {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
	} else {
		return nil, fmt.Errorf("invalid private key type: %s", block.Type)
	}

	key.Precompute()

	if err := key.Validate(); err != nil {
		return nil, err
	}

	return
}

// GetToken returns the token for IAP
func GetToken(saPath, clientID string) (token string, err error) {
	sa, err := ioutil.ReadFile(saPath)
	if err != nil {
		return
	}
	conf, err := google.JWTConfigFromJSON(sa)
	if err != nil {
		return
	}
	rsaKey, _ := readRsaPrivateKey(conf.PrivateKey)
	iat := time.Now()
	exp := iat.Add(time.Hour)
	jwt := &jws.ClaimSet{
		Iss: conf.Email,
		Aud: TokenURI,
		Iat: iat.Unix(),
		Exp: exp.Unix(),
		PrivateClaims: map[string]interface{}{
			"target_audience": clientID,
		},
	}
	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	msg, err := jws.Encode(jwsHeader, jwt, rsaKey)
	if err != nil {
		return
	}

	v := neturl.Values{}
	v.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	v.Set("assertion", msg)

	ctx := context.Background()
	hc := oauth2.NewClient(ctx, nil)
	resp, err := hc.PostForm(TokenURI, v)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		IDToken     string `json:"id_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return token, err
	}

	token = tokenRes.IDToken
	return
}
