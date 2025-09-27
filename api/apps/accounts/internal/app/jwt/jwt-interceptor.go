package jwt

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/api/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Claims struct {
	UserId       string `json:"userId"`
	UserEmail    string `json:"userEmail"`
	UserName     string `json:"userName"`
	UserVerified bool   `json:"userVerified"`
	jwt.RegisteredClaims
}

type JWTInterceptor struct {
	jwtSecret string
}

func NewJWTInterceptor(jwtSecret string) *JWTInterceptor {
	return &JWTInterceptor{
		jwtSecret: jwtSecret,
	}
}

// UnaryServerInterceptor returns a server interceptor function to validate JWT tokens
func (j *JWTInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip authentication for internal methods
		if strings.Contains(info.FullMethod, "Internal") {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata not found")
		}

		// Get authorization header
		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization header not found")
		}

		// Extract token from "Bearer <token>" format
		authHeader := authHeaders[0]
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate JWT token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(j.jwtSecret), nil
		})

		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token claims")
		}

		// Add user info to metadata for downstream use
		newMD := metadata.New(map[string]string{
			"userId":       claims.UserId,
			"userEmail":    claims.UserEmail,
			"userName":     claims.UserName,
			"userVerified": fmt.Sprintf("%t", claims.UserVerified),
		})
		ctx = metadata.NewIncomingContext(ctx, metadata.Join(md, newMD))

		// Call the handler
		return handler(ctx, req)
	}
}

// ExtractUserContext extracts user information from context metadata
func ExtractUserContext(ctx context.Context) (userId, userEmail, userName string, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", "", "", errors.New("metadata not found")
	}

	userIds := md.Get("userId")
	if len(userIds) == 0 {
		return "", "", "", errors.New("userId not found in metadata")
	}

	userEmails := md.Get("userEmail")
	if len(userEmails) == 0 {
		return "", "", "", errors.New("userEmail not found in metadata")
	}

	userNames := md.Get("userName")
	if len(userNames) == 0 {
		userNames = []string{""}
	}

	return userIds[0], userEmails[0], userNames[0], nil
}