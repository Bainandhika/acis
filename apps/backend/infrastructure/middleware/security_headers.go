package middleware

import (
	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds OWASP recommended security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Protect against clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable browser XSS filtering
		c.Header("X-XSS-Protection", "1; mode=block")

		// Control referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions policy to restrict sensitive browser features
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Content Security Policy
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: https:; connect-src 'self' *;")

		// Enforce HTTPS in production via HSTS
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		c.Next()
	}
}
