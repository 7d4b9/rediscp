package main_test

import (
	"fmt"
	"github.com/go-redis/redis"
	. "github.com/jobteaser/redis-copy"
	"sort"
	"testing"
)

func seedDB(c *redis.Client) {
	c.FlushDB()
	for i := 0; i < 100000; i++ {
		c.Set(fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i), 0)
	}
}

func TestCopyDB(t *testing.T) {
	c1 := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	seedDB(c1)

	c1.LPush("foo", "bar", "baz")

	c1.HSet("users:1", "username", "gearnode")
	c1.HSet("users:1", "age", 18)

	c1.SAdd("blabla", 1, 2, 3, 4, "foo", "bar")

	c1.ZAdd(
		"bloblo",
		redis.Z{Score: 1, Member: "bloblo"},
		redis.Z{Score: 2, Member: 1},
		redis.Z{Score: 3, Member: 2},
		redis.Z{Score: 4, Member: 4},
		redis.Z{Score: 5, Member: "foo"},
		redis.Z{Score: 6, Member: "bar"},
	)

	c2 := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: "",
		DB:       0,
	})
	c2.FlushDB()

	if CopyDB(c1, c2) != nil {
		t.FailNow()
	}

	for i := 0; i < 100000; i++ {
		v, err := c2.Get(fmt.Sprintf("k%d", i)).Result()
		if err != nil {
			t.Error("error get key")
		}

		if v != fmt.Sprintf("v%d", i) {
			t.Error("key is not valid => ", v)
			t.FailNow()
		}
	}

	{
		if c2.LLen("foo").Val() != 2 {
			t.FailNow()
		}

		comp := c2.LRange("foo", 0, -1).Val()
		expected := []string{"baz", "bar"}

		for i := range comp {
			if comp[i] != expected[i] {
				t.Error("bad list order")
			}
		}

		user := c2.HGetAll("users:1").Val()
		if user["username"] != "gearnode" {
			t.Error("hash not imported")
		}
		if user["age"] != "18" {
			t.Error("hash not imported")
		}
	}

	{
		expected := c1.SMembers("blabla").Val()
		comp := c2.SMembers("blabla").Val()

		sort.Strings(expected)
		sort.Strings(comp)

		for i := range expected {
			if comp[i] != expected[i] {
				t.Error("set is not properly imported")
			}
		}
	}

	{
		expected := c1.ZRangeWithScores("bloblo", 0, -1).Val()
		comp := c2.ZRangeWithScores("bloblo", 0, -1).Val()

		for i := range expected {
			if expected[i].Member != comp[i].Member {
				t.Error("sorted set not properly imported")
			}
			if expected[i].Score != comp[i].Score {
				t.Error("sorted set not properly imported")
			}
		}
	}

}
