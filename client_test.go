package rbns

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestClientCheck(t *testing.T) {
	ctx := context.Background()
	client, err := Connection(ctx, WithHost(fmt.Sprintf("%s:%d", "notification-management-envoy", 10000)), WithDialOption(grpc.WithInsecure()))
	if assert.NoError(t, err) {
		r, err := client.Check("user1", "default", "create:test")
		assert.NoError(t, err)
		assert.True(t, r)
	}
}
