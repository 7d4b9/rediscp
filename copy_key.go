package main

import (
	"github.com/go-redis/redis"
	"sync"
)

// CopyKey dump key in the redis source and dump it in the redis destination
//
// NOTE: This func does not use DUMP and RESTORE command provide by Redis, because these commands
// use RDB format. This format can include breaking change. For example to downgrade Redis, the RDB
// file is not compatible.
func CopyKey(src *redis.Client, dest *redis.Client, key string, err chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	t := src.Type(key).Val()
	switch t {
	case "string":
		ttl := src.TTL(key).Val()
		v := src.Get(key).Val()
		if ttl < 0 {
			ttl = 0
		}
		_, e := dest.Set(key, v, ttl).Result()
		if e != nil {
			err <- &errRedisCp{e.Error(), key, string(ttl), t}
		}
	case "list":
		v := src.LRange(key, 0, -1).Val()
		x := make([]interface{}, len(v))
		for i, v := range v {
			x[i] = v
		}
		_, e := dest.RPush(key, x...).Result()
		if e != nil {
			err <- &errRedisCp{e.Error(), key, "0", t}
		}
	case "hash":
		v := src.HGetAll(key).Val()
		for i, j := range v {
			_, e := dest.HSet(key, i, j).Result()
			if e != nil {
				err <- &errRedisCp{e.Error(), key, "0", t}
			}
		}
	case "set":
		v := src.SMembers(key).Val()
		x := make([]interface{}, len(v))
		for i, v := range v {
			x[i] = v
		}
		_, e := dest.SAdd(key, x...).Result()
		if e != nil {
			err <- &errRedisCp{e.Error(), key, "0", t}
		}
	case "zset":
		v := src.ZRangeWithScores(key, 0, -1).Val()
		_, e := dest.ZAdd(key, v...).Result()
		if e != nil {
			err <- &errRedisCp{e.Error(), key, "0", t}
		}
	default:
		err <- &errRedisCp{"Unsupported type", key, "0", t}
	}
}
