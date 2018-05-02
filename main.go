package main

import (
	"os"

	"github.com/go-redis/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func main() {
	log.SetOutput(os.Stdout)

	var (
		redisDest string
		redisSrc  string
	)

	rootCmd := &cobra.Command{
		Use:   "rediscp",
		Short: "Copy all keys in the redis db (by default 0)",
		Long:  "A simple cli to copy an db redis to another redis db",
		Args:  cobra.NoArgs,
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
				for _, err := range errs {
					if err, ok := err.(*errRedisCp); ok {
						log.WithFields(log.Fields{"key": err.Key(), "ttl": err.TTL(), "key_type": err.Type()}).Warn(err.Error())
					} else {
						log.Warn(err)
					}
				}
				os.Exit(1)
			}
		},
	}

	rootCmd.Flags().StringVarP(&redisSrc, "src", "s", "", "Source redis instance to copy from (required)")
	rootCmd.Flags().StringVarP(&redisDest, "dest", "d", "", "Destination redis instance to paste in (required)")
	rootCmd.MarkFlagRequired("src")
	rootCmd.MarkFlagRequired("dest")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
