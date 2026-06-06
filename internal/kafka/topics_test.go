package kafka

import "testing"

func TestTopicString(t *testing.T) {
	if got := TopicOrderPlaced.String(); got != "orders.placed" {
		t.Fatalf("expected orders.placed, got %q", got)
	}
}
