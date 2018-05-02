package main

import (
	"github.com/go-redis/redis"
	"sync"
)

// CopyDB copy all the keys in the Redis DB
func CopyDB(src *redis.Client, dest *redis.Client) []error {
	var errs []error
	ch := make(chan error)
	wg := sync.WaitGroup{}

	iter := src.Scan(0, "", 0).Iterator()

	for iter.Next() {
		wg.Add(1)
		go CopyKey(src, dest, iter.Val(), ch, &wg)
	}

	wg.Wait()
	close(ch)

	for err := range ch {
		errs = append(errs, err)
	}

	return errs
}
