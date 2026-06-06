package kafka

type Topic string

const (
	TopicOrderPlaced Topic = "orders.placed"
)

func (t Topic) String() string {
	return string(t)
}
