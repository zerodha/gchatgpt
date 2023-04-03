package gchat

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const (
	jwkRoot    string = "https://www.googleapis.com/service_accounts/v1/jwk/"
	chatIssuer string = "chat@system.gserviceaccount.com"
)

var (
	refreshInterval = 30 * time.Minute
)

// JWKVerifier is a wrapper around the jwk.Cache.
type JWKVerifier struct {
	c   *jwk.Cache
	url string
}

// NewJWKVerifier creates a new JWKVerifier.
func NewJWKVerifier() (*JWKVerifier, error) {
	ctx := context.Background()
	c := jwk.NewCache(ctx)

	if err := c.Register(jwkRoot+chatIssuer,
		jwk.WithRefreshInterval(refreshInterval)); err != nil {
		return nil, err
	}

	_, err := c.Refresh(ctx, jwkRoot+chatIssuer)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh google JWKS: %s\n", err)
	}

	return &JWKVerifier{
		c:   c,
		url: jwkRoot + chatIssuer,
	}, nil
}

// VerifyJWT verifies the JWT token.
func (jv *JWKVerifier) VerifyJWT(audience, tokenRaw string) error {
	token, err := jwt.Parse(tokenRaw, func(token *jwt.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid")
		}

		ctx := context.Background()
		jwSet, err := jv.c.Get(ctx, jv.url)
		if err != nil {
			return nil, fmt.Errorf("failed to get google JWKS: %s\n", err)
		}

		if key, ok := jwSet.LookupKeyID(kid); ok {
			var pubkey interface{}
			if err := key.Raw(&pubkey); err != nil {
				return nil, err
			}

			return pubkey, nil
		}

		return nil, errors.New("failed to find key")
	})

	if err != nil {
		return err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return errors.New("failed to parse claims")
	}

	for key, value := range claims {
		if key == "aud" && value == audience {
			return nil
		}
	}

	return errors.New("failed to authenticate token")
}
