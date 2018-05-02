package main

import (
	"errors"
	"os"
	"sync"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CopyKey dump and restore key
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
			log.WithFields(log.Fields{"key": key, "ttl": ttl, "err": e.Error()}).Error("failed to dump the redis key")
			err <- e
		}
	case "list":
		v := src.LRange(key, 0, -1).Val()
		x := make([]interface{}, len(v))
		for i, v := range v {
			x[i] = v
		}
		_, e := dest.RPush(key, x...).Result()
		if e != nil {
			log.WithFields(log.Fields{"key": key, "err": e.Error()}).Error("failed to dump the redis key")
			err <- e
		}
	case "hash":
		v := src.HGetAll(key).Val()
		for i, j := range v {
			_, e := dest.HSet(key, i, j).Result()
			if e != nil {
				log.WithFields(log.Fields{"key": key, "err": e.Error()}).Error("failed to dump the redis key")
				err <- e
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
			log.WithFields(log.Fields{"key": key, "err": e.Error()}).Error("failed to dump the redis key")
			err <- e
		}
	case "zset":
		v := src.ZRangeWithScores(key, 0, -1).Val()
		_, e := dest.ZAdd(key, v...).Result()
		if e != nil {
			log.WithFields(log.Fields{"key": key, "err": e.Error()}).Error("failed to dump the redis key")
			err <- e
		}
	default:
		err <- errors.New("Unsupported type:" + t)
	}
}

// CopyDB perform the database data copy
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

var (
	rootCmd = &cobra.Command{
		Use:   "redis-copy",
		Short: "Dump redis to another redis",
		Long:  `A simple wrapper to copy cloud redis data to an other cloud redis instance`,
	}

	redisDest string
	redisSrc  string

	copyCmd = &cobra.Command{
		Use:   "copy",
		Short: "Copy all redis data in an other redis instance",
		Run: func(cmd *cobra.Command, args []string) {
			srcOps, err := redis.ParseURL(redisSrc)
			if err != nil {
				log.Fatal(err)
			}

			destOps, err := redis.ParseURL(redisDest)
			if err != nil {
				log.Fatal(err)
			}

			srcClient := redis.NewClient(srcOps)
			if _, err := srcClient.Ping().Result(); err != nil {
				log.Fatal(err)
			}

			destClient := redis.NewClient(destOps)
			if _, err := destClient.Ping().Result(); err != nil {
				log.Fatal(err)
			}

			errs := CopyDB(srcClient, destClient)
			if errs != nil {
				log.Fatal(errs)
			}
		},
	}
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().StringVarP(&redisSrc, "src", "s", "", "Source redis instance to copy from (required)")
	copyCmd.Flags().StringVarP(&redisDest, "dest", "d", "", "Destination redis instance to paste in (required)")
	copyCmd.MarkFlagRequired("src")
	copyCmd.MarkFlagRequired("dest")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
