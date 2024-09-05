package redis

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pool          *redis.Pool        // Redis connection pool
	redisAddr     = "127.0.0.1:6379" //redis address Host:port
	redisPassword = "123456"         // Redis password
)

// init
func Initialize() *redis.Pool {
	return &redis.Pool{
		MaxIdle:     50, // max idle amount
		MaxActive:   30, //max active number, (0 means no limit)
		IdleTimeout: 300 * time.Second,
		Dial: func() (redis.Conn, error) {
			// 1. Open connection
			connection, err := redis.Dial("tcp", redisAddr)
			if err != nil {
				fmt.Println("Cannot connect to redis: ", err)
				return nil, err
			}
			// 2. authenticate
			if _, err = connection.Do("AUTH", redisPassword); err != nil {
				fmt.Println("Cannot authenticate to redis: ", err)
				connection.Close()
				return nil, err
			}
			return connection, nil
		},
		TestOnBorrow: func(conn redis.Conn, lastUsed time.Time) error { // test connection
			// idle time < 60s => dont check
			if time.Since(lastUsed) < time.Minute {
				return nil
			}
			// we send PING to check connection
			_, err := conn.Do("PING")
			return err
		},
	}
}

func init() {
	pool = Initialize()
	// get all keys
	data, err := pool.Get().Do("KEYS", "*")
	if err != nil {
		fmt.Println("Error fetching all keys: ", err)
	}
	fmt.Println("All keys in redis: ", data)
}

func RedisPool() *redis.Pool {
	return pool
}
