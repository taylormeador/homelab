package server

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()

	resp, err := handler(ctx, req)

	code := status.Code(err)
	log.Printf("method=%s status=%s duration=%s", info.FullMethod, code, time.Since(start))

	return resp, err
}
