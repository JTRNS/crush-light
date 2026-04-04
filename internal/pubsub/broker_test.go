package pubsub

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBrokerPublishAndCancelSubscription(t *testing.T) {
	t.Parallel()

	broker := NewBroker[int]()
	ctx, cancel := context.WithCancel(t.Context())
	sub := broker.Subscribe(ctx)

	require.Equal(t, 1, broker.GetSubscriberCount())

	broker.Publish(CreatedEvent, 42)

	select {
	case event := <-sub:
		require.Equal(t, CreatedEvent, event.Type)
		require.Equal(t, 42, event.Payload)
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for published event")
	}

	cancel()

	require.Eventually(t, func() bool {
		return broker.GetSubscriberCount() == 0
	}, time.Second, 10*time.Millisecond)
}

func TestBrokerShutdownClosesSubscribers(t *testing.T) {
	t.Parallel()

	broker := NewBroker[string]()
	sub := broker.Subscribe(t.Context())

	broker.Shutdown()

	_, ok := <-sub
	require.False(t, ok)

	closedSub := broker.Subscribe(t.Context())
	_, ok = <-closedSub
	require.False(t, ok)
}

func TestNewBrokerWithOptions_UsesConfiguredBufferSize(t *testing.T) {
	t.Parallel()

	broker := NewBrokerWithOptions[int](1, 1000)
	sub := broker.Subscribe(t.Context())

	broker.Publish(CreatedEvent, 1)
	broker.Publish(CreatedEvent, 2)
	broker.Publish(CreatedEvent, 3)

	var got []int
	for {
		select {
		case event := <-sub:
			got = append(got, event.Payload)
		default:
			require.Equal(t, []int{1}, got)
			return
		}
	}
}
