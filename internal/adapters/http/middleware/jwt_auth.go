// internal/adapters/http/middleware/jwt_auth.go
package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kristianrpo/document-management-microservice/internal/adapters/http/dto/response/shared"
)

// UserClaims representa las claims del JWT del microservicio de auth
type UserClaims struct {
	IDCitizen int64  `json:"id_citizen"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

type contextKey string

const UserContextKey contextKey = "user_claims"
const OAuthContextKey contextKey = "oauth_claims"

// JWTAuthMiddleware valida el JWT token del microservicio de auth
type JWTAuthMiddleware struct {
	jwtSecret []byte
}

func NewJWTAuthMiddleware(jwtSecret string) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		jwtSecret: []byte(jwtSecret),
	}
}

// Authenticate middleware that validates JWT token
func (m *JWTAuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("JWT auth: missing Authorization header from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "missing authorization header"))
			c.Abort()
			return
		}

		// Expected format: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("JWT auth: invalid authorization header format: %q from %s", authHeader, c.ClientIP())
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			// Log parse error for debugging (do not log tokens in prod)
			log.Printf("JWT parse error: %v, token: %s", err, tokenString)
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("INVALID_TOKEN", "invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*UserClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("INVALID_TOKEN", "invalid token claims"))
			c.Abort()
			return
		}

		// Log useful claim fields for debugging (user id, citizen id, email, role)
		// IDCitizen is int64; use %d. There is no separate user_id string, so log id_citizen once.
		log.Printf("JWT validated - id_citizen=%d email=%s role=%s", claims.IDCitizen, claims.Email, claims.Role)

		// Store claims in context for handlers to use
		c.Set(string(UserContextKey), claims)
		c.Next()
	}
}

// RequireRole middleware that checks if user has required role
func (m *JWTAuthMiddleware) RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetUserFromContext(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "user not authenticated"))
			c.Abort()
			return
		}

		// Check if user has one of the allowed roles
		hasRole := false
		for _, role := range allowedRoles {
			if claims.Role == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, shared.NewErrorResponse("FORBIDDEN", "insufficient permissions"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// OAuthTokenClaims representa las claims de un token de tipo client_credentials
type OAuthTokenClaims struct {
	ClientID string   `json:"client_id"`
	Scopes   []string `json:"scopes"`
	TokenID  string   `json:"jti"`
	IssuedAt int64    `json:"iat"`
	ExpireAt int64    `json:"exp"`
	Type     string   `json:"type"`
	jwt.RegisteredClaims
}

// AuthenticateClient validates a client credentials JWT and stores OAuthTokenClaims in context
func (m *JWTAuthMiddleware) AuthenticateClient() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Printf("JWT auth: missing Authorization header from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "missing authorization header"))
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Printf("JWT auth: invalid authorization header format: %q from %s", authHeader, c.ClientIP())
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "invalid authorization header format"))
			c.Abort()
			return
		}

		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &OAuthTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return m.jwtSecret, nil
		})

		if err != nil {
			log.Printf("OAuth JWT parse error: %v", err)
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("INVALID_TOKEN", "invalid or expired token"))
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*OAuthTokenClaims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("INVALID_TOKEN", "invalid token claims"))
			c.Abort()
			return
		}

		if claims.Type != "client_credentials" {
			c.JSON(http.StatusForbidden, shared.NewErrorResponse("FORBIDDEN", "token is not client_credentials type"))
			c.Abort()
			return
		}

		log.Printf("OAuth token validated - client_id=%s token_id=%s scopes=%v", claims.ClientID, claims.TokenID, claims.Scopes)

		c.Set(string(OAuthContextKey), claims)
		c.Next()
	}
}

// RequireClientCredentials checks that an OAuth client_credentials token is present and optionally validates scopes
func (m *JWTAuthMiddleware) RequireClientCredentials(requiredScopes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value, exists := c.Get(string(OAuthContextKey))
		if !exists {
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("UNAUTHORIZED", "client credentials token not provided"))
			c.Abort()
			return
		}

		claims, ok := value.(*OAuthTokenClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, shared.NewErrorResponse("INVALID_TOKEN", "invalid token claims"))
			c.Abort()
			return
		}

		if len(requiredScopes) > 0 {
			scopeSet := make(map[string]struct{}, len(claims.Scopes))
			for _, s := range claims.Scopes {
				scopeSet[s] = struct{}{}
			}
			for _, rs := range requiredScopes {
				if _, ok := scopeSet[rs]; !ok {
					c.JSON(http.StatusForbidden, shared.NewErrorResponse("FORBIDDEN", "insufficient scopes"))
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// GetUserFromContext retrieves user claims from gin context
func GetUserFromContext(c *gin.Context) *UserClaims {
	value, exists := c.Get(string(UserContextKey))
	if !exists {
		return nil
	}

	claims, ok := value.(*UserClaims)
	if !ok {
		return nil
	}

	return claims
}

// GetUserIDCitizen helper to get user's citizen ID from context
func GetUserIDCitizen(c *gin.Context) (int64, error) {
	claims := GetUserFromContext(c)
	if claims == nil {
		return 0, fmt.Errorf("user not authenticated")
	}
	return claims.IDCitizen, nil
}

// Context helpers for standard context.Context (not gin)
func GetUserFromStandardContext(ctx context.Context) *UserClaims {
	value := ctx.Value(UserContextKey)
	if value == nil {
		return nil
	}

	claims, ok := value.(*UserClaims)
	if !ok {
		return nil
	}

	return claims
}

func SetUserInStandardContext(ctx context.Context, claims *UserClaims) context.Context {
	return context.WithValue(ctx, UserContextKey, claims)
}
