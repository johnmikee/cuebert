package lifecycle

import "sync"

var ( // lifecycle is a global variable to hold the lifecycle client.
	lifecycle *Client
	rs        chan struct{}
	startsig  chan struct{}
	stopsig   chan struct{}
	once      sync.Once
)

// LifeCycle interface remains unchanged.
type LifeCycle interface {
	Start()
	Stop()
	Update()
}

// Client is a struct to hold the signals for the client.
type Client struct {
	reloadSignal   chan struct{}
	startSignal    chan struct{}
	stopSignal     chan struct{}
	isRunningFunc  func() bool
	setRunningFunc func(bool)
}

// Start sends a signal to start the client.
func (c *Client) Start() {
	if c.setRunningFunc != nil && !c.isRunningFunc() {
		c.setRunningFunc(true)
		c.startSignal <- struct{}{}
	}
}

// Stop sends a signal to stop the client.
func (c *Client) Stop() {
	if c.setRunningFunc != nil && c.isRunningFunc() {
		c.stopSignal <- struct{}{}
		c.setRunningFunc(false)
	}
}

// Update sends a signal to reload the client.
func (c *Client) Update() {
	c.reloadSignal <- struct{}{}
}

// IsRunning returns true if the client is running.
func (c *Client) IsRunning() bool {
	if c.isRunningFunc != nil {
		return c.isRunningFunc()
	}
	return false
}

// SetRunningFunc sets the functions to check if the client is running and to set the running state.
func (c *Client) SetRunningFunc(isRunning func() bool, setRunning func(bool)) {
	c.isRunningFunc = isRunning
	c.setRunningFunc = setRunning
}

// ReloadSignal returns the reload signal channel.
func (c *Client) ReloadSignal() chan struct{} {
	return c.reloadSignal
}

// StartSignal returns the start signal channel.
func (c *Client) StartSignal() chan struct{} {
	return c.startSignal
}

// StopSignal returns the stop signal channel.
func (c *Client) StopSignal() chan struct{} {
	return c.stopSignal
}

// New returns a new lifecycle client.
func New() *Client {
	once.Do(func() {
		rs = make(chan struct{})
		startsig = make(chan struct{})
		stopsig = make(chan struct{})

		lifecycle = &Client{
			reloadSignal: rs,
			startSignal:  startsig,
			stopSignal:   stopsig,
		}
	})
	return lifecycle
}
