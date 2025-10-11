package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterConfig holds rate limiting configuration
type RateLimiterConfig struct {
	Requests int           // Number of requests allowed
	Period   time.Duration // Time period for the limit
}

// Default rate limiting configurations
var (
	// Login rate limiting: 5 attempts per minute
	LoginRateLimit = RateLimiterConfig{
		Requests: 5,
		Period:   time.Minute,
	}
	
	// Registration rate limiting: 3 attempts per minute
	RegistrationRateLimit = RateLimiterConfig{
		Requests: 3,
		Period:   time.Minute,
	}
	
	// General API rate limiting: 100 requests per minute
	APIRateLimit = RateLimiterConfig{
		Requests: 100,
		Period:   time.Minute,
	}
)

// RateLimiter creates a rate limiting middleware
func RateLimiter(config RateLimiterConfig) gin.HandlerFunc {
	// Create a rate limiter instance
	store := memory.NewStore()
	rate := limiter.Rate{
		Period: config.Period,
		Limit:  int64(config.Requests),
	}
	
	instance := limiter.New(store, rate)
	
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()
		
		// Get rate limit context
		context, err := instance.Get(c, clientIP)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Rate limit error"})
			c.Abort()
			return
		}
		
		// Set rate limit headers
		c.Header("X-RateLimit-Limit", string(rune(context.Limit)))
		c.Header("X-RateLimit-Remaining", string(rune(context.Remaining)))
		c.Header("X-RateLimit-Reset", string(rune(context.Reset)))
		
		// Check if rate limit exceeded
		if context.Reached {
			resetTime := time.Unix(context.Reset, 0)
			retryAfter := int(time.Until(resetTime).Seconds())
			if retryAfter < 0 {
				retryAfter = 0
			}
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
				"retry_after": retryAfter,
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// LoginRateLimiter returns rate limiter for login endpoints
func LoginRateLimiter() gin.HandlerFunc {
	return RateLimiter(LoginRateLimit)
}

// RegistrationRateLimiter returns rate limiter for registration endpoints
func RegistrationRateLimiter() gin.HandlerFunc {
	return RateLimiter(RegistrationRateLimit)
}

// APIRateLimiter returns rate limiter for general API endpoints
func APIRateLimiter() gin.HandlerFunc {
	return RateLimiter(APIRateLimit)
}
