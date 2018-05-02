package main

// ErrRedisCp represent the standard error
type errRedisCp struct {
	err string
	key string
	ttl string
	typ string
}

func (e *errRedisCp) Error() string {
	return e.err
}

// Key of the object
func (e *errRedisCp) Key() string {
	return e.key
}

// TTL of the object
func (e *errRedisCp) TTL() string {
	return e.ttl
}

// Type of the object
func (e *errRedisCp) Type() string {
	return e.typ
}
