package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/nikimonax/go-metrics/internal/app/interfaces"
)

type PublisherMock struct {
	mock.Mock
}

// Publish implements [interfaces.EventPublisher].
func (publisher *PublisherMock) Publish(
	ctx context.Context,
	event interfaces.Event,
) {
	publisher.Called(ctx, event)
}

var _ interfaces.EventPublisher = (*PublisherMock)(nil)
