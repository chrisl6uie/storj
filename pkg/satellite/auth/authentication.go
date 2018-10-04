// Copyright (C) 2018 Storj Labs, Inc.
// See LICENSE for copying information.

package auth

import (
	"context"
	"crypto/ecdsa"
	"encoding/base64"
	"strings"

	"github.com/gtank/cryptopasta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"storj.io/storj/pkg/peertls"
	"storj.io/storj/pkg/pointerdb/auth"
	"storj.io/storj/pkg/provider"
)

// ResponseGenerator interface for generating signature
type ResponseGenerator interface {
	Generate(ctx context.Context) error
}

// NewSatelliteAuthenticator creates instance of satellite authenticator
func NewSatelliteAuthenticator(generator ResponseGenerator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{},
		info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{},
		err error) {

		// TODO(michal) we can extend auth to overlay requests also
		if strings.HasPrefix(info.FullMethod, "/pointerdb") {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok || !auth.ValidateAPIKey(strings.Join(md["apikey"], "")) {
				return nil, status.Errorf(codes.Unauthenticated, "Invalid API credential")
			}

			err := generator.Generate(ctx)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "%v", err)
			}
		}

		return handler(ctx, req)
	}
}

type defaultResponseGenerator struct {
	identity *provider.FullIdentity
}

// NewResponseGenerator creates default response generator based on identity
func NewResponseGenerator(identity *provider.FullIdentity) ResponseGenerator {
	return &defaultResponseGenerator{identity: identity}
}

func (s *defaultResponseGenerator) Generate(ctx context.Context) error {
	pk, ok := s.identity.Key.(*ecdsa.PrivateKey)
	if !ok {
		return peertls.ErrUnsupportedKey.New("%T", pk)
	}

	signature, err := cryptopasta.Sign(s.identity.ID.Bytes(), pk)
	if err != nil {
		return err
	}

	encoding := base64.StdEncoding
	encodedSignature := encoding.EncodeToString(signature)
	err = grpc.SetHeader(ctx, metadata.Pairs("signature", encodedSignature))
	if err != nil {
		return err
	}

	return nil
}