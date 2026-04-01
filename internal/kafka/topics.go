package kafka

type Topic string

const (
	TopicOrderCreated Topic = "orders.created"
	TopicPayments     Topic = "payments"
	TopicDeliveries   Topic = "deliveries"
)

func (t Topic) String() string {
	return string(t)
}
