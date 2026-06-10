package kafka

import "testing"

func TestTopicString(t *testing.T) {
	if got := TopicOrderCreated.String(); got != "orders.created" {
		t.Fatalf("expected orders.placed, got %q", got)
	}
}
