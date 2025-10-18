package dbagent

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"civicweave/backend/pkg/metadb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// APIKeyHeader is the gRPC metadata key for API key
	APIKeyHeader = "x-api-key"
	// ClientIDHeader is the gRPC metadata key for client ID
	ClientIDHeader = "x-client-id"
	// RequestIDHeader is the gRPC metadata key for request ID
	RequestIDHeader = "x-request-id"
)

// AuthInterceptor provides API key authentication for gRPC services
type AuthInterceptor struct {
	metaRepo *metadb.Repository
}

// NewAuthInterceptor creates a new authentication interceptor
func NewAuthInterceptor(metaRepo *metadb.Repository) *AuthInterceptor {
	return &AuthInterceptor{
		metaRepo: metaRepo,
	}
}

// UnaryServerInterceptor returns a gRPC unary server interceptor that validates API keys
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// Extract API key from metadata
		apiKeys := md.Get(APIKeyHeader)
		if len(apiKeys) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing API key")
		}

		clientIDs := md.Get(ClientIDHeader)
		if len(clientIDs) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing client ID")
		}

		apiKey := apiKeys[0]
		clientID := clientIDs[0]

		// Validate API key
		valid, err := a.validateAPIKey(clientID, apiKey)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "authentication error: %v", err)
		}

		if !valid {
			return nil, status.Errorf(codes.Unauthenticated, "invalid API key")
		}

		// Add authenticated context
		ctx = context.WithValue(ctx, "client_id", clientID)
		ctx = context.WithValue(ctx, "api_key", apiKey)

		// Extract request ID for audit logging
		requestIDs := md.Get(RequestIDHeader)
		if len(requestIDs) > 0 {
			ctx = context.WithValue(ctx, "request_id", requestIDs[0])
		}

		// Continue with the request
		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a gRPC stream server interceptor that validates API keys
func (a *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// Extract API key from metadata
		apiKeys := md.Get(APIKeyHeader)
		if len(apiKeys) == 0 {
			return status.Errorf(codes.Unauthenticated, "missing API key")
		}

		clientIDs := md.Get(ClientIDHeader)
		if len(clientIDs) == 0 {
			return status.Errorf(codes.Unauthenticated, "missing client ID")
		}

		apiKey := apiKeys[0]
		clientID := clientIDs[0]

		// Validate API key
		valid, err := a.validateAPIKey(clientID, apiKey)
		if err != nil {
			return status.Errorf(codes.Internal, "authentication error: %v", err)
		}

		if !valid {
			return status.Errorf(codes.Unauthenticated, "invalid API key")
		}

		// Create authenticated context
		ctx := ss.Context()
		ctx = context.WithValue(ctx, "client_id", clientID)
		ctx = context.WithValue(ctx, "api_key", apiKey)

		// Extract request ID for audit logging
		requestIDs := md.Get(RequestIDHeader)
		if len(requestIDs) > 0 {
			ctx = context.WithValue(ctx, "request_id", requestIDs[0])
		}

		// Wrap the server stream with authenticated context
		authenticatedStream := &authenticatedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Continue with the request
		return handler(srv, authenticatedStream)
	}
}

// validateAPIKey validates the API key against the metadata database
func (a *AuthInterceptor) validateAPIKey(clientID, apiKey string) (bool, error) {
	// Hash the API key for comparison
	hash := sha256.Sum256([]byte(apiKey))
	publicKeyHash := base64.StdEncoding.EncodeToString(hash[:])

	// Validate against metadata database
	return a.metaRepo.ValidateAPIKey(clientID, publicKeyHash)
}

// authenticatedServerStream wraps a grpc.ServerStream with authenticated context
type authenticatedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the authenticated context
func (s *authenticatedServerStream) Context() context.Context {
	return s.ctx
}

// ClientAuth provides authentication utilities for gRPC clients
type ClientAuth struct {
	ClientID string
	APIKey   string
}

// NewClientAuth creates a new client authentication helper
func NewClientAuth(clientID, apiKey string) *ClientAuth {
	return &ClientAuth{
		ClientID: clientID,
		APIKey:   apiKey,
	}
}

// GetClientMetadata returns the metadata for gRPC client calls
func (c *ClientAuth) GetClientMetadata() metadata.MD {
	return metadata.New(map[string]string{
		APIKeyHeader:    c.APIKey,
		ClientIDHeader:  c.ClientID,
		RequestIDHeader: generateRequestID(),
	})
}

// WithAuth adds authentication metadata to a context
func (c *ClientAuth) WithAuth(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, c.GetClientMetadata())
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	return fmt.Sprintf("req_%d_%s", time.Now().UnixNano(), randomString(8))
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// AuditLogger provides audit logging functionality
type AuditLogger struct {
	metaRepo *metadb.Repository
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(metaRepo *metadb.Repository) *AuditLogger {
	return &AuditLogger{
		metaRepo: metaRepo,
	}
}

// LogRequest logs an audit entry for a gRPC request
func (a *AuditLogger) LogRequest(
	ctx context.Context,
	action string,
	databaseID *string,
	deploymentID *string,
	statusCode int,
	errorMessage string,
	executionTimeMs int,
	requestSizeBytes int,
	responseSizeBytes int,
	metadata map[string]interface{},
) error {
	// Extract client information from context
	clientID, _ := ctx.Value("client_id").(string)
	requestID, _ := ctx.Value("request_id").(string)

	// Get client IP from context (if available)
	clientIP := getClientIP(ctx)

	// Create audit log entry
	entry := &metadb.AuditLog{
		DatabaseID:        databaseID,
		DeploymentID:      deploymentID,
		Action:            action,
		UserAgent:         fmt.Sprintf("db-client/%s", clientID),
		ClientIP:          clientIP,
		RequestID:         requestID,
		StatusCode:        statusCode,
		ErrorMessage:      errorMessage,
		ExecutionTimeMs:   executionTimeMs,
		RequestSizeBytes:  requestSizeBytes,
		ResponseSizeBytes: responseSizeBytes,
		CreatedAt:         time.Now(),
		Metadata:          marshalMetadata(metadata),
	}

	return a.metaRepo.LogAuditEntry(entry)
}

// getClientIP extracts client IP from gRPC context
func getClientIP(ctx context.Context) string {
	// Try to get IP from gRPC peer info
	if peer, ok := ctx.Value("peer").(string); ok {
		// Extract IP from peer address (format: "ipv4:127.0.0.1:12345")
		parts := strings.Split(peer, ":")
		if len(parts) >= 2 {
			return parts[1]
		}
	}

	// Try to get IP from metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if xForwardedFor := md.Get("x-forwarded-for"); len(xForwardedFor) > 0 {
			// X-Forwarded-For can contain multiple IPs, take the first one
			ips := strings.Split(xForwardedFor[0], ",")
			return strings.TrimSpace(ips[0])
		}
		if xRealIP := md.Get("x-real-ip"); len(xRealIP) > 0 {
			return xRealIP[0]
		}
	}

	return "unknown"
}

// marshalMetadata converts metadata map to JSON
func marshalMetadata(metadata map[string]interface{}) []byte {
	if len(metadata) == 0 {
		return []byte("{}")
	}

	// Simple JSON marshaling for metadata
	var parts []string
	for key, value := range metadata {
		parts = append(parts, fmt.Sprintf(`"%s": "%v"`, key, value))
	}
	jsonStr := fmt.Sprintf("{%s}", strings.Join(parts, ","))
	return []byte(jsonStr)
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// Allow checks if a request from the given client should be allowed
func (r *RateLimiter) Allow(clientID string) bool {
	now := time.Now()
	cutoff := now.Add(-r.window)

	// Clean old requests
	if requests, exists := r.requests[clientID]; exists {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}
		r.requests[clientID] = validRequests
	} else {
		r.requests[clientID] = []time.Time{}
	}

	// Check if limit exceeded
	if len(r.requests[clientID]) >= r.limit {
		return false
	}

	// Add current request
	r.requests[clientID] = append(r.requests[clientID], now)
	return true
}

// RateLimitInterceptor provides rate limiting for gRPC services
func (r *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Get client ID from context
		clientID, ok := ctx.Value("client_id").(string)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing client ID")
		}

		// Check rate limit
		if !r.Allow(clientID) {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
		}

		// Continue with the request
		return handler(ctx, req)
	}
}
