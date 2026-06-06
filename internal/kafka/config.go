package kafka

type ConsumerConfig struct {
	Brokers        []string
	Topic          string
	GroupID        string
	Partition      int
	MinBytes       int
	MaxBytes       int
	CommitInterval int
	StartOffset    int64
}
