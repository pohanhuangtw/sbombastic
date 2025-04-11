package messaging

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

const sbombasticSubject = "sbombastic"

func prePareRoutes() string {
	selfName, _ := os.Hostname()
	routeHosts := []string{}

	for i := 0; i < 3; i++ {
		peer := fmt.Sprintf("sbombastic-controller-%d", i)
		if peer == selfName {
			continue
		}
		host := fmt.Sprintf("nats://%s.sbombastic-nats-cluster.sbombastic.svc.cluster.local:6222", peer)
		routeHosts = append(routeHosts, host)
	}
	return strings.Join(routeHosts, ",")
}

func NewServer() (*server.Server, error) {
	serverName, _ := os.Hostname()
	fmt.Println("serverName", serverName)
	opts := &server.Options{
		ServerName: serverName, // 從環境變數中獲取唯一的伺服器名稱
		JetStream:  true,
		Cluster: server.ClusterOpts{
			Name: "sbombastic",
			Host: "0.0.0.0",
			Port: 6222,
			// 在實際部署中，這應該是動態發現的
		},
		Routes: server.RoutesFromStr(prePareRoutes()),
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS server: %w", err)
	}
	ns.ConfigureLogger()

	go ns.Start()

	if !ns.ReadyForConnections(20 * time.Second) {
		return nil, fmt.Errorf("NATS server not ready for connections: %w", err)
	}

	return ns, nil
}

func NewJetStreamContext(ns *server.Server) (nats.JetStreamContext, error) {
	nc, err := nats.Connect(ns.ClientURL())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return js, nil
}

func AddStream(js nats.JetStreamContext, storage nats.StorageType) error {
	_, err := js.AddStream(&nats.StreamConfig{
		Name: "SBOMBASTIC",
		// We use WorkQueuePolicy to ensure that each message is removed once it is processed.
		Retention: nats.WorkQueuePolicy,
		Subjects:  []string{sbombasticSubject},
		Storage:   storage,
	})
	if err != nil {
		return fmt.Errorf("failed to add JetStream stream: %w", err)
	}

	return nil
}

func NewSubscription(url, durable string) (*nats.Subscription, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS server: %w", err)
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256))
	if err != nil {
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	sub, err := js.PullSubscribe(sbombasticSubject, durable, nats.InactiveThreshold(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to JetStream stream: %w", err)
	}

	return sub, nil
}
