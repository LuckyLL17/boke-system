package service

import (
	"testing"

	"podcast-platform/internal/domain"
)

// Subscribe -> FindByEmail -> IncrementSubscribers -> shouldIncrementSubscriberCount
func TestActiveSubscriptionDoesNotIncreaseChannelCount(t *testing.T) {
	existing := &domain.Subscriber{Status: 1}
	if shouldIncrementSubscriberCount(existing) {
		panic("active subscription was treated as a new subscriber")
	}
}
