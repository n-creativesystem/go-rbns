package rbns

import (
	"context"
	"n-creativesystem/go-rbns-sdk/proto"

	"google.golang.org/grpc"
)

type PermissionClient interface {
	Check(userKey, organizationId string, permissionNames ...string) (bool, error)
}

type permissionClient struct {
	client proto.PermissionClient
	ctx    context.Context
}

func newPermission(con *grpc.ClientConn, ctx context.Context) PermissionClient {
	return &permissionClient{
		client: proto.NewPermissionClient(con),
		ctx:    ctx,
	}
}

func (c *permissionClient) Check(userKey, organizationName string, permissionNames ...string) (bool, error) {
	ps := make([]string, len(permissionNames))
	copy(ps, permissionNames)
	res, err := c.client.Check(c.ctx, &proto.PermissionCheckRequest{
		UserKey:          userKey,
		OrganizationName: organizationName,
		PermissionNames:  ps,
	})
	if err != nil {
		return false, err
	}
	return res.Result, nil
}
