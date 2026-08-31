package server

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// requestID assigns a unique request identifier to each incoming request,
// mirroring chi's middleware.RequestID. It reuses an inbound X-Request-Id
// header when present and echoes the id back on the response.
func requestID(c *gin.Context) {
	id := c.GetHeader("X-Request-Id")
	if id == "" {
		id = generateRequestID()
	}
	c.Writer.Header().Set("X-Request-Id", id)
	c.Next()
}

func generateRequestID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "req"
	}
	return hex.EncodeToString(b)
}

// realIP sets the client's real IP from X-Forwarded-For / X-Real-IP headers,
// mirroring chi's middleware.RealIP. This keeps request.RemoteAddr consistent
// with the original behavior for handlers that inspect it.
func realIP(c *gin.Context) {
	if ip := realIPFromHeaders(c); ip != "" {
		c.Request.RemoteAddr = ip
	}
	c.Next()
}

func realIPFromHeaders(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		ip := strings.TrimSpace(parts[0])
		if net.ParseIP(ip) != nil {
			return ip
		}
	}
	if xrip := c.GetHeader("X-Real-Ip"); xrip != "" {
		if net.ParseIP(xrip) != nil {
			return xrip
		}
	}
	return ""
}
