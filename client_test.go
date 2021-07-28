package rbns

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientCheck(t *testing.T) {
	ctx := context.Background()
	client, _ := Connection(ctx, WithHost(fmt.Sprintf("%s:%d", "api-rbac-dev", 8888)), WithApiKey("5d78ced0-c6a0-471d-90f3-a0ec7653172e"))
	r, err := client.Check("user1", "default", "create:test")
	assert.NoError(t, err)
	assert.True(t, r)
}
