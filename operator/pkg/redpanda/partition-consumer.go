package redpanda

//
// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"sync"
//
// 	"github.com/twmb/franz-go/pkg/kgo"
// 	"github.com/kloudlite/operator/types/errors"
// 	"github.com/kloudlite/operator/types/logging"
// )
//
// // Implementation derived from https://github.com/twmb/franz-go/blob/master/examples/goroutine_per_partition_consuming/manual_commit/main.go#L26
//
// type pConsumer struct {
// 	cl        *kgo.Client
// 	topic     string
// 	partition int32
// 	logger    logging.Logger
//
// 	quit    chan struct{}
// 	done    chan struct{}
// 	records chan []*kgo.Record
// }
//
// func (pc *pConsumer) commitRecord(ctx context.Context, record *kgo.Record) {
// 	if err := pc.cl.CommitRecords(ctx, record); err != nil {
// 		pc.logger.Infof(
// 			"failed commiting to kafka (offset: %d, topic: %s  partition: %d), error: %v\n", record.Offset,
// 			pc.topic, pc.partition, err.Error(),
// 		)
// 	}
// 	pc.logger.Infof("committed to kafka, (offset: %d, topic: %s, partition: %d)\n", record.Offset, pc.topic, pc.partition)
// }
//
// func (pc *pConsumer) consume(reader ReaderFunc) {
// 	if reader == nil {
// 		return
// 	}
// 	defer close(pc.done)
// 	fmt.Printf("Starting consume for  t %s p %d\n", pc.topic, pc.partition)
// 	defer fmt.Printf("Closing consume for t %s p %d\n", pc.topic, pc.partition)
// 	for {
// 		select {
// 		case <-pc.quit:
// 			return
// 		case records := <-pc.records:
// 			for _, record := range records {
// 				if record == nil {
// 					pc.commitRecord(context.Background(), record)
// 				}
//
// 				// TODO: need to have retry logic for failed messages
// 				if err := reader(
// 					KafkaMessage{
// 						Key:        record.Key,
// 						Value:      record.Value,
// 						Timestamp:  record.Timestamp,
// 						Topic:      record.Topic,
// 						Partition:  record.Partition,
// 						ProducerId: record.ProducerID,
// 						Offset:     record.Offset,
// 					},
// 				); err != nil {
// 					// TODO: (here also)
// 					pc.commitRecord(context.TODO(), record)
// 				}
// 				pc.commitRecord(context.TODO(), record)
// 			}
// 		}
// 	}
// }
//
// type tp struct {
// 	t string
// 	p int32
// }
//
// type splitter struct {
// 	client         *kgo.Client
// 	pConsumers     map[tp]*pConsumer
// 	logger         logging.Logger
// 	maxRetries     int
// 	maxPollRecords int
// }
//
// func (s *splitter) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
// 	for topic, partitions := range assigned {
// 		for _, partition := range partitions {
// 			pc := &pConsumer{
// 				cl:        cl,
// 				topic:     topic,
// 				partition: partition,
// 				logger:    s.logger,
//
// 				quit:    make(chan struct{}),
// 				done:    make(chan struct{}),
// 				records: make(chan []*kgo.Record),
// 			}
// 			s.pConsumers[tp{topic, partition}] = pc
// 		}
// 	}
// }
//
// func (s *splitter) lost(_ context.Context, _ *kgo.Client, lost map[string][]int32) {
// 	var wg sync.WaitGroup
// 	defer wg.Wait()
//
// 	for topic, partitions := range lost {
// 		for _, partition := range partitions {
// 			tp := tp{topic, partition}
// 			pc := s.pConsumers[tp]
// 			delete(s.pConsumers, tp)
// 			close(pc.quit)
// 			fmt.Printf("waiting for work to finish t %s p %d\n", topic, partition)
// 			wg.Add(1)
// 			go func() { <-pc.done; wg.Done() }()
// 		}
// 	}
// }
//
// func (s *splitter) Ping(ctx context.Context) error {
// 	return s.client.Ping(ctx)
// }
//
// func (s *splitter) Close() {
// 	for i := range s.pConsumers {
// 		close(s.pConsumers[i].quit)
// 	}
// 	s.client.Close()
// }
//
// func (s *splitter) StartConsuming(reader ReaderFunc) error {
// 	if reader == nil {
// 		return errors.Newf("no reader defined for message consumption")
// 	}
//
// 	for i := range s.pConsumers {
// 		go s.pConsumers[i].consume(reader)
// 	}
//
// 	for {
// 		fetches := s.client.PollRecords(context.Background(), s.maxPollRecords)
// 		if fetches.IsClientClosed() {
// 			return errors.Newf("consumer client is closed")
// 		}
// 		fetches.EachError(
// 			func(s string, i int32, err error) {
// 				fmt.Println("[ERR]: ", s, i, err)
// 			},
// 		)
// 		fetches.EachPartition(
// 			func(p kgo.FetchTopicPartition) {
// 				pConsumer := s.pConsumers[tp{p.Topic, p.Partition}]
// 				pConsumer.records <- p.Records
// 			},
// 		)
// 		s.client.AllowRebalance()
// 	}
// }
//
// func NewConsumer(brokers string, group, topic string, consumerOpts *ConsumerOpts) (consumer, error) {
// 	cOpts := consumerOpts.getWithDefaults()
//
// 	s := &splitter{
// 		pConsumers:     make(map[tp]*pConsumer),
// 		maxRetries:     *cOpts.MaxRetries,
// 		maxPollRecords: *cOpts.MaxPollRecords,
// 	}
//
// 	opts := []kgo.Opt{
// 		kgo.SeedBrokers(strings.Split(brokers, ",")...),
// 		kgo.ConsumerGroup(group),
// 		kgo.ConsumeTopics(topic),
// 		kgo.OnPartitionsAssigned(s.assigned),
// 		kgo.OnPartitionsRevoked(s.lost),
// 		kgo.OnPartitionsLost(s.lost),
// 		kgo.DisableAutoCommit(),
// 		// kgo.BlockRebalanceOnPoll(),
// 	}
//
// 	client, err := kgo.NewClient(opts...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	s.client = client
// 	return s, nil
// }
