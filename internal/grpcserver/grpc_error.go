package grpcserver

import (
	"errors"
	"net/http"

	httperr "github.com/coolycow/gophkeeper/internal/error"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// grpcError переводит CustomError (и прочие ошибки) в gRPC status.
func grpcError(err error) error {
	if err == nil {
		return nil
	}

	var ce httperr.CustomError
	if errors.As(err, &ce) {
		switch ce.StatusCode {
		case http.StatusBadRequest: // 400
			return status.Error(codes.InvalidArgument, ce.Message)
		case http.StatusNotFound: // 404
			return status.Error(codes.NotFound, ce.Message)
		case http.StatusConflict: // 409 → AlreadyExists (текст — полный short URL, как в HTTP)
			return status.Error(codes.AlreadyExists, ce.Message)
		case http.StatusUnauthorized: // 401
			return status.Error(codes.Unauthenticated, ce.Message)
		case http.StatusUnsupportedMediaType: // 415
			return status.Error(codes.InvalidArgument, ce.Message)
		case http.StatusGone: // 410
			return status.Error(codes.FailedPrecondition, ce.Message)
		default:
			return status.Error(codes.Internal, ce.Message)
		}
	}

	return status.Errorf(codes.Internal, "%v", err)
}
