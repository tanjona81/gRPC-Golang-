package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ctxKey string

const RequestIDKey ctxKey = "request_id"

func TimeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Create a new context that expires after 'timeout'
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Create a channel to capture the result of the handler
		// (Handler runs in its own goroutine so we can "interrupt" it)
		ch := make(chan struct {
			resp any
			err  error
		}, 1)

		go func() {
			resp, err := handler(ctx, req)
			ch <- struct {
				resp any
				err  error
			}{resp, err}
		}()

		// Wait for either the handler to finish OR the context to timeout
		select {
		case res := <-ch:
			return res.resp, res.err
		case <-ctx.Done():
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded by server timeout")
		}
	}
}

func RequestIDInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	var requestID string

	// Try to get ID from incoming metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ids := md.Get("x-request-id")
		if len(ids) > 0 {
			requestID = ids[0]
		}
	}

	// If no ID exists, generate a new one
	if requestID == "" {
		requestID = uuid.New().String()
	}

	// Store the ID in the context for downstream use
	ctx = context.WithValue(ctx, RequestIDKey, requestID)

	// Also send it back to the client in the header (Response Metadata)
	header := metadata.Pairs("x-request-id", requestID)
	grpc.SendHeader(ctx, header)

	return handler(ctx, req)
}

func LoggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()

	// Execute the handler
	resp, err := handler(ctx, req)

	// Pull the ID we just set in the previous interceptor
	requestID, _ := ctx.Value(RequestIDKey).(string)

	// Calculate metrics
	duration := time.Since(start)
	st, _ := status.FromError(err)

	// Calculate level based on status code
	level := slog.LevelInfo
	if err != nil {
		level = slog.LevelWarn
	}

	// Structured Logging
	slog.Log(ctx, level, "gRPC Request",
		"method", info.FullMethod,
		"duration_ms", duration.Milliseconds(),
		"code", st.Code().String(),
		"error", err,
		"request_id", requestID,
	)

	return resp, err
}
