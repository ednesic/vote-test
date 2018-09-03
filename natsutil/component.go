package natsutil

import (
	"fmt"
	"sync"

	"github.com/nats-io/go-nats-streaming"
	"github.com/nats-io/nuid"
)

type StreamingComponent interface {
	ConnectToNATSStreaming(clusterID string, options ...stan.Option) error
	NATS() stan.Conn
	ID() string
	Name() string
	Shutdown() error
}

// StreamingComponent is contains reusable logic related to handling
// of the connection to NATS Streaming in the system.
type streamingComponent struct {
	// cmu is the lock from the component.
	cmu sync.Mutex

	// id is a unique identifier used for this component.
	id string

	// nc is the connection to NATS Streaming.
	nc stan.Conn

	// kind is the type of component.
	kind string
}

// NewStreamingComponent creates a StreamingComponent
func NewStreamingComponent(kind string) StreamingComponent {
	id := nuid.Next()
	return &streamingComponent{
		id:   id,
		kind: kind,
	}
}

// ConnectToNATSStreaming connects to NATS Streaming
func (c *streamingComponent) ConnectToNATSStreaming(clusterID string, options ...stan.Option) error {
	c.cmu.Lock()
	defer c.cmu.Unlock()

	// Connect to NATS with Cluster Id, Client Id and customized options.
	nc, err := stan.Connect(clusterID, c.id, options...)
	if err != nil {
		return err
	}
	c.nc = nc
	return err
}

// NATS returns the current NATS connection.
func (c *streamingComponent) NATS() stan.Conn {
	c.cmu.Lock()
	defer c.cmu.Unlock()
	return c.nc
}

// ID returns the ID from the component.
func (c *streamingComponent) ID() string {
	c.cmu.Lock()
	defer c.cmu.Unlock()
	return c.id
}

// Name is the label used to identify the NATS connection.
func (c *streamingComponent) Name() string {
	c.cmu.Lock()
	defer c.cmu.Unlock()
	return fmt.Sprintf("%s:%s", c.kind, c.id)
}

// Shutdown makes the component go away.
func (c *streamingComponent) Shutdown() error {
	c.NATS().Close()
	return nil
}
