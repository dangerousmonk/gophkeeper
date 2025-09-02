package middleware

import (
	"context"
	"log/slog"
	"strings"

	"github.com/dangerousmonk/gophkeeper/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey struct {
	name string
}

var (
	userIDContextKey = &contextKey{"userID"}
)

const (
	bearerPrefix = "Bearer "
	headerName   = "authorization"
)

// PublicMethods is a set of gRPC methods that don't require authentication.
var PublicMethods = map[string]struct{}{
	"/server.GophKeeper/RegisterUser": {},
	"/server.GophKeeper/LoginUser":    {},
	"/server.GophKeeper/Ping":         {},
}

// IsPublicMethod checks if the given gRPC full method is in the public methods list.
func IsPublicMethod(fullMethod string) bool {
	_, ok := PublicMethods[fullMethod]
	return ok
}

// AuthUnaryInterceptor reads JWT token from metadata and validates it.
func AuthUnaryInterceptor(jwtManager utils.Authenticator) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		if IsPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		// Retrieve metaData from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			slog.Warn("AuthUnaryInterceptor:no md", slog.Any("context", ctx))
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}

		// Retrieve token
		values := md.Get(headerName)
		if len(values) == 0 || !strings.HasPrefix(values[0], bearerPrefix) {
			return nil, status.Errorf(codes.Unauthenticated, "missing or invalid token")
		}

		token := strings.TrimPrefix(values[0], bearerPrefix)

		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token")
		}
		userID := claims.UserID

		ctx = context.WithValue(ctx, userIDContextKey, userID)
		return handler(ctx, req)
	}
}

// UserIDFromContext retrieves the authenticated user's ID from the context.
// Returns:
//   - string: The user ID if found and of correct type
//   - bool:   True if user ID was found and is valid string, false otherwise
func UserIDFromContext(ctx context.Context) (int, bool) {
	if ctx == nil {
		return -1, false
	}

	if val := ctx.Value(userIDContextKey); val != nil {
		if userID, ok := val.(int); ok {
			return userID, true
		}
	}
	return -1, false
}

// wrappedServerStream wraps the original ServerStream to override the context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedServerStream) Context() context.Context {
	return s.ctx
}

// StreamAuthInterceptor is uses to check user token for for streaming RPCs
func StreamAuthInterceptor(jwtManager utils.Authenticator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		userID, err := authenticate(ctx, jwtManager)
		if err != nil {
			slog.Warn("StreamAuthInterceptor: authentication failed", slog.Any("error", err))
			return status.Errorf(codes.Unauthenticated, "authentication failed: %v", err)
		}

		ctx = context.WithValue(ctx, userIDContextKey, userID)

		wrappedStream := &wrappedServerStream{ss, ctx}
		slog.Info("StreamAuthInterceptor: user authenticated", slog.Int("user_id", userID))
		return handler(srv, wrappedStream)
	}
}

// authenticate extracts and validates token from gRPC metadata
func authenticate(ctx context.Context, jwtManager utils.Authenticator) (int, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return -1, status.Errorf(codes.Unauthenticated, "missing metadata")
	}

	authHeaders := md.Get(headerName)
	if len(authHeaders) == 0 {
		return -1, status.Errorf(codes.Unauthenticated, "missing authorization header")
	}

	if len(authHeaders) == 0 || !strings.HasPrefix(authHeaders[0], bearerPrefix) {
		slog.Warn("authenticate: invalid authorization format", slog.Any("authHeaders", authHeaders))
		return -1, status.Errorf(codes.Unauthenticated, "invalid authorization format")
	}

	token := strings.TrimPrefix(authHeaders[0], bearerPrefix)
	if token == "" {
		slog.Warn("authenticate: empty tokent", slog.Any("authHeaders", authHeaders))
		return -1, status.Errorf(codes.Unauthenticated, "empty token")
	}

	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		slog.Warn("authenticate: invalid token", slog.String("token", token))
		return -1, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	userID := claims.UserID
	return userID, nil
}
