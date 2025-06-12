package client

import (
	"context"
	"github.com/Killazius/linkify-proto/pkg/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	api api.AuthClient
	log *zap.SugaredLogger
}

func NewAuthClient(
	log *zap.SugaredLogger,
	addr string,
) (*Client, error) {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &Client{
		api: api.NewAuthClient(cc),
		log: log,
	}, nil
}

func (c *Client) ValidateToken(ctx context.Context, in *api.TokenRequest, opts ...grpc.CallOption) (*api.TokenResponse, error) {
	resp, err := c.api.ValidateToken(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
