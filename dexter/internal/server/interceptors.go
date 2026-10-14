package server

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type roleKey struct{}

func RoleFromContext(ctx context.Context) string {
	role, _ := ctx.Value(roleKey{}).(string)
	return role
}

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	code := status.Code(err)
	log.Printf("method=%s status=%s duration=%s", info.FullMethod, code, time.Since(start))

	return resp, err
}

func NewAuthInterceptor(adminKey, demoKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "no metadata")
		}

		keys := md.Get("authorization")
		if len(keys) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "no api key")
		}

		key := keys[0]
		var role string
		switch key {
		case adminKey:
			role = "admin"
		case demoKey:
			role = "demo"
		default:
			return nil, status.Errorf(codes.Unauthenticated, "invalid api key")
		}

		ctx = context.WithValue(ctx, roleKey{}, role)
		return handler(ctx, req)
	}
}
