package gapi

import (
	"context"
	"fmt"

	db "github.com/adwait-godbole/go-bank/db/sqlc"
	"github.com/adwait-godbole/go-bank/pb"
	"github.com/adwait-godbole/go-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) VerifyEmail(ctx context.Context, req *pb.VerifyEmailRequest) (*pb.VerifyEmailResponse, error) {
	violations := validateVerifyEmailRequest(req)
	if violations != nil {
		return nil, invalidArgumentError(violations)
	}

	txResult, err := s.store.VerifyEmailTx(ctx, db.VerifyEmailTxParams{
		Id:         req.GetId(),
		SecretCode: req.GetSecretCode(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to verify email")
	}

	res := &pb.VerifyEmailResponse{
		IsVerified: txResult.User.IsEmailVerified,
	}
	return res, nil
}

func validateVerifyEmailRequest(req *pb.VerifyEmailRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.GetId() <= 0 {
		violations = append(violations, fieldViolation("id", fmt.Errorf("must be a positive integer")))
	}

	if err := val.ValidateString(req.GetSecretCode(), 32, 128); err != nil {
		violations = append(violations, fieldViolation("secret_code", err))
	}

	return
}
