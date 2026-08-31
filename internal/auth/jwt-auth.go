package auth

import (
	"context"
	"go-chi-sqlite-jwt-starter/config"
	"go-chi-sqlite-jwt-starter/internal/models"
	"go-chi-sqlite-jwt-starter/internal/provider"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret []byte

func InitializeTokenVerifier() {
	secret := config.Variables.AUTH_PRIVATE_KEY
	jwtSecret = []byte(secret)
}

// UseAuthMiddleware registers the JWT verification pipeline on the given group.
// It mirrors the original chi setup: seek/verify/validate the token, reject
// invalid tokens with 401, then load the user from the database and place it on
// the request context (403 if the account no longer exists).
func UseAuthMiddleware(r *gin.RouterGroup) {
	r.Use(verifierMiddleware())
	r.Use(myAuthMiddleware())
}

// verifierMiddleware seeks a bearer token, verifies its HS256 signature and
// validates its claims. On any failure it aborts with 401, matching the
// original token verification behavior.
func verifierMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := tokenFromRequest(c.Request)
		if tokenString == "" {
			http.Error(c.Writer, "no token found", http.StatusUnauthorized)
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(c.Writer, "invalid token", http.StatusUnauthorized)
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(c.Writer, "invalid token", http.StatusUnauthorized)
			c.Abort()
			return
		}

		ctx := context.WithValue(c.Request.Context(), claimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// myAuthMiddleware loads the authenticated user from the database using the
// user_id claim and stores it on the request context, so downstream handlers
// can read it via utils.GetUserFromContext. Preserves the original 403 message.
func myAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Request.Context().Value(claimsContextKey).(jwt.MapClaims)

		userID := int64(claims["user_id"].(float64))
		user, err := provider.Provider.UserService.GetUser(userID)
		if err != nil {
			http.Error(c.Writer, "You are authenticated, but we could not find your account.", http.StatusForbidden)
			c.Abort()
			return
		}
		ctx := context.WithValue(c.Request.Context(), models.ContextKeys.User, user)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GenerateJWT(user models.User) (string, error) {
	ttl := 7 * 24 * time.Hour // 1 week
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"sub":     user.Username,
		"exp":     time.Now().UTC().Add(ttl).Unix(),
	})
	return token.SignedString(jwtSecret)
}

type authContextKey string

const claimsContextKey authContextKey = "jwt_claims"

// tokenFromRequest extracts a bearer token from the Authorization header,
// falling back to the "jwt" cookie, matching jwtauth's default token finders.
func tokenFromRequest(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		parts := strings.SplitN(h, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return strings.TrimSpace(parts[1])
		}
	}
	if cookie, err := r.Cookie("jwt"); err == nil {
		return cookie.Value
	}
	return ""
}
