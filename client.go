package rbns

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
)

const (
	ClientKey = "rbns-client/v1.0.0"
)

var defaultConfig = &config{
	dialOptions: []grpc.DialOption{},
	host:        "localhost:6565",
}

type config struct {
	dialOptions []grpc.DialOption
	apiKey      string
	host        string
}

type Option func(conf *config)

func WithDialOption(opt ...grpc.DialOption) Option {
	return func(conf *config) {
		conf.dialOptions = append(conf.dialOptions, opt...)
	}
}

func WithApiKey(apiKey string) Option {
	return func(conf *config) {
		conf.apiKey = fmt.Sprintf("Bearer %s", apiKey)
	}
}

func WithHost(host string) Option {
	return func(conf *config) {
		conf.host = host
	}
}

type Client struct {
	con  *grpc.ClientConn
	ctx  context.Context
	conf config
}

func (c *Client) Close() error {
	if c.con != nil {
		return c.con.Close()
	}
	return nil
}

func (c *Client) permission() PermissionClient {
	return newPermission(c.con, c.ctx)
}

func (c *Client) Check(userKey, organizationName string, permissionNames ...string) (bool, error) {
	return c.permission().Check(userKey, organizationName, permissionNames...)
}

func Connection(ctx context.Context, opts ...Option) (*Client, error) {
	conf := &config{}
	*conf = *defaultConfig
	for _, opt := range opts {
		opt(conf)
	}
	if conf.apiKey != "" {
		md := metadata.New(map[string]string{"authorization": conf.apiKey})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}
	con, err := grpc.DialContext(ctx, conf.host, conf.dialOptions...)
	if err != nil {
		return nil, err
	}
	resp, err := healthpb.NewHealthClient(con).Check(ctx, &healthpb.HealthCheckRequest{
		Service: "",
	})
	if err != nil {
		return nil, fmt.Errorf("Status RPC failure: %s", err.Error())
	}
	if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
		return nil, fmt.Errorf("Status unhealthy: %s", resp.GetStatus().String())
	}
	return &Client{
		con:  con,
		ctx:  ctx,
		conf: *conf,
	}, nil
}
