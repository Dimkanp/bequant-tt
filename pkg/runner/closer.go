package runner

type Closer interface {
	Close() error
}

type closer struct {
	c      chan error
	closer Closer
}

func (c *closer) Run() error {
	return <-c.c
}

func (c *closer) Stop() error {
	close(c.c)
	return c.closer.Close()
}

// FromCloser creates Runner that has effect only on Stop() method
// calling c.Close() method
func FromCloser(c Closer) *closer {
	closer := &closer{
		c:      make(chan error),
		closer: c,
	}

	return closer
}
