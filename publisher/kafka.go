package publisher

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/IBM/sarama"
	"github.com/goccy/go-json"

	"github.com/ihippik/wal-listener/v2/config"
)

// KafkaPublisher represent event publisher with Kafka broker.
type KafkaPublisher struct {
	producer sarama.SyncProducer
}

// NewKafkaPublisher return new KafkaPublisher instance.
func NewKafkaPublisher(producer sarama.SyncProducer) *KafkaPublisher {
	return &KafkaPublisher{producer: producer}
}

func (p *KafkaPublisher) Publish(s string, event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, _, err = p.producer.SendMessage(prepareMessage(s, data))

	return err
}

// NewProducer return new Kafka producer instance.
func NewProducer(pCfg *config.PublisherCfg) (sarama.SyncProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Return.Successes = true

	if pCfg.EnableTLS {
		tlsCfg, err := newTLSCfg(pCfg.ClientCert, pCfg.ClientKey, pCfg.CACert)
		if err != nil {
			return nil, err
		}

		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = tlsCfg
	}

	producer, err := sarama.NewSyncProducer([]string{pCfg.Address}, cfg)

	return producer, err
}

// prepareMessage prepare message for Kafka producer.
func prepareMessage(topic string, data []byte) *sarama.ProducerMessage {
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Partition: -1,
		Value:     sarama.ByteEncoder(data),
	}

	return msg
}

func newTLSCfg(certFile, keyFile, caCert string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	ca, err := os.ReadFile(caCert)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)
	cfg.RootCAs = caCertPool

	return cfg, nil
}