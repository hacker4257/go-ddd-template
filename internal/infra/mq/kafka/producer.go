package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/hacker4257/go-ddd-template/internal/domain/event"
)

type Producer struct {
	cl *kgo.Client
}

func NewProducer(brokers []string) (*Producer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		// 先最小配置，后续你可以加：压缩、幂等、重试、超时等
	)
	if err != nil {
		return nil, fmt.Errorf("kgo new client: %w", err)
	}
	return &Producer{cl: cl}, nil
}

func (p *Producer) Publish(ctx context.Context, topic string, e event.Event, headers map[string]string) error {
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}

	var hs []kgo.RecordHeader
	for k, v := range headers {
		hs = append(hs, kgo.RecordHeader{Key: k, Value: []byte(v)})
	}

	rec := &kgo.Record{
		Topic:   topic,
		Key:     []byte(e.Key),
		Value:   b,
		Headers: hs,
	}

	// 同步等待 ack（简单可靠；后续性能优化可改异步+回调）
	res := p.cl.ProduceSync(ctx, rec)
	return res.FirstErr()
}

func (p *Producer) Close() error {
	p.cl.Close()
	return nil
}
