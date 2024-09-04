package contexter_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"loyalty/internal/lib/contexter"
)

func TestGetRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), contexter.RequestID, "test-request-id")

	requestID := contexter.GetRequestID(ctx)

	assert.Equal(t, "test-request-id", requestID)
}
