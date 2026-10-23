// Package auth implements shared OpenID Connect JWT authentication.
package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CakeForKit/rsoi-lab2/lb-core/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const usernameKey = "authenticated_username"

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KID string `json:"kid"`
	KTY string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var keyCache struct {
	sync.Mutex
	keys    map[string]*rsa.PublicKey
	expires time.Time
}

// Middleware validates a bearer JWT and makes its user identity available in
// the Gin context. OAuth endpoints and health checks remain public.
func Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if isPublic(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}
		username, err := validateBearer(ctx.GetHeader("Authorization"))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "unauthorized"})
			return
		}
		ctx.Set(usernameKey, username)
		ctx.Next()
	}
}

// Username returns the identity obtained from a validated JWT.
func Username(ctx *gin.Context) string {
	username, _ := ctx.Get(usernameKey)
	result, _ := username.(string)
	return result
}

func isPublic(path string) bool {
	return path == "/api/v1/authorize" || path == "/api/v1/callback" || strings.HasPrefix(path, "/manage/")
}

func validateBearer(header string) (string, error) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", fmt.Errorf("bearer token is required")
	}
	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(parts[1], claims, keyFunc,
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg(), jwt.SigningMethodRS384.Alg(), jwt.SigningMethodRS512.Alg()}),
		jwt.WithIssuer(config.CoreConfig.Auth.Issuer),
	)
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}
	claim := config.CoreConfig.Auth.UserClaim
	if claim == "" {
		claim = "sub"
	}
	username, _ := claims[claim].(string)
	if username == "" {
		return "", fmt.Errorf("user claim is missing")
	}
	return username, nil
}

func keyFunc(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, fmt.Errorf("token key id is missing")
	}
	keys, err := getKeys()
	if err != nil {
		return nil, err
	}
	key := keys[kid]
	if key == nil {
		return nil, fmt.Errorf("unknown token key")
	}
	return key, nil
}

func getKeys() (map[string]*rsa.PublicKey, error) {
	keyCache.Lock()
	defer keyCache.Unlock()
	if time.Now().Before(keyCache.expires) && keyCache.keys != nil {
		return keyCache.keys, nil
	}
	if config.CoreConfig.Auth.JWKSURL == "" {
		return nil, fmt.Errorf("OIDC_JWKS_URL is not configured")
	}
	response, err := (&http.Client{Timeout: 5 * time.Second}).Get(config.CoreConfig.Auth.JWKSURL)
	if err != nil {
		return nil, fmt.Errorf("load JWKs: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("load JWKs: %s", response.Status)
	}
	var set jwks
	if err = json.NewDecoder(response.Body).Decode(&set); err != nil {
		return nil, fmt.Errorf("decode JWKs: %w", err)
	}
	keys := make(map[string]*rsa.PublicKey, len(set.Keys))
	for _, item := range set.Keys {
		if item.KTY != "RSA" || item.KID == "" {
			continue
		}
		key, parseErr := rsaKey(item)
		if parseErr == nil {
			keys[item.KID] = key
		}
	}
	if len(keys) == 0 {
		return nil, fmt.Errorf("JWKs contains no RSA keys")
	}
	ttl := config.CoreConfig.Auth.JWKSCacheTTL
	if ttl <= 0 {
		ttl = 3600
	}
	keyCache.keys = keys
	keyCache.expires = time.Now().Add(time.Duration(ttl) * time.Second)
	return keys, nil
}

func rsaKey(key jwk) (*rsa.PublicKey, error) {
	n, err := base64.RawURLEncoding.DecodeString(key.N)
	if err != nil {
		return nil, err
	}
	e, err := base64.RawURLEncoding.DecodeString(key.E)
	if err != nil {
		return nil, err
	}
	value := 0
	for _, byteValue := range e {
		value = value<<8 | int(byteValue)
	}
	if value == 0 {
		return nil, fmt.Errorf("invalid RSA exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(n), E: value}, nil
}
