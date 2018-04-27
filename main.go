package main

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// CopyKey dump and restore key
func CopyKey(src *redis.Client, dest *redis.Client, key string, err chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	ttl := src.TTL(key).Val()
	dump, e := src.Dump(key).Result()
	if e != nil {
		log.WithFields(log.Fields{"key": key, "ttl": ttl}).Error("failed to dump the redis key")
		err <- e
		return
	}

	if ttl < 0 {
		ttl = 0
	}

	_, e = dest.RestoreReplace(key, ttl, dump).Result()
	if e != nil {
		log.WithFields(log.Fields{"key": key, "ttl": ttl}).Error("failed to restore the redis key")
		err <- e
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

// CheckMemorySpace check if dest redis have require memory space
func CheckMemorySpace(src *redis.Client, dest *redis.Client) bool {
	info, err := src.Info("memory").Result()
	if err != nil {
		return false
	}
	result := make(map[string]string)
	for _, match := range RedisMemoryMatch.FindAllString(info, -1) {
		part := strings.Split(match, ":")
		result[part[0]] = part[1]
	}
	info2, err := dest.Info("memory").Result()
	if err != nil {
		return false
	}
	result2 := make(map[string]string)
	for _, match := range RedisMemoryMatch.FindAllString(info2, -1) {
		part := strings.Split(match, ":")
		result2[part[0]] = part[1]
	}

	i, err := strconv.Atoi(result2["used_memory"])
	if err != nil {
		panic(err)
	}
	j, err := strconv.Atoi(result2["total_system_memory"])
	if err != nil {
		panic(err)
	}
	k, err := strconv.Atoi(result["used_memory"])
	if err != nil {
		panic(err)
	}
	if (j - i) < k {
		return false
	}
	return true
}

var (
	// RedisMemoryMatch represent regex to match redis memory informations
	RedisMemoryMatch = regexp.MustCompile(`(used_memory:(?P<used>\d+)|total_system_memory:(?P<total>\d+))`)

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
				os.Exit(1)
			}
			destOps, err := redis.ParseURL(redisDest)
			if err != nil {
				log.Fatal(err)
				os.Exit(1)
			}
			srcClient := redis.NewClient(srcOps)
			destClient := redis.NewClient(destOps)
			if !CheckMemorySpace(srcClient, destClient) {
				log.Error("the destination redis does not have require memory space")
				os.Exit(1)
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
