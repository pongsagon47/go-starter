package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/unrolled/secure" // Or another security library
)

// Helmet provides a security middleware with recommended HTTP headers.
func Helmet() gin.HandlerFunc {
	// Configure the security middleware with your desired settings.
	// You should customize these based on your application's needs.
	secureMiddleware := secure.New(secure.Options{
		// Recommended options
		FrameDeny:          true,
		ContentTypeNosniff: true,
		BrowserXssFilter:   true,
		ReferrerPolicy:     "strict-origin-when-cross-origin",

		// HSTS is important for HTTPS
		// HSTS is usually a must-have for production environments
		// The max-age should be set to a long duration, e.g., one year (31536000 seconds)
		// and includeSubdomains should be true if applicable
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,

		// CSP is powerful but can be complex to configure.
		// Be very careful with this. Start with a report-only mode first.
		// ContentSecurityPolicy: "default-src 'self'",
	})

	return func(c *gin.Context) {
		err := secureMiddleware.Process(c.Writer, c.Request)
		if err != nil {
			c.AbortWithStatus(500)
			return
		}
		c.Next()
	}
}
