package gorad

import (
	"time"
	"github.com/garyburd/redigo/redis"
	"github.com/xiaojunli/gorad/config"
	"github.com/xiaojunli/gorad/logger"
)

var pool *redis.Pool

/**
 * 获取客户端
 */
func GetRedis() redis.Conn {
	if nil == pool {
		logger.Debug("ooo - redis pool is not ready, init pool...")
		pool = initRedisPool()
		logger.Debug("ooo - redis pool init success...")
	}

	return pool.Get()
}

/**
 * 初始化了连接池
 */
func initRedisPool() (*redis.Pool) {
	// 从配置文件获取redis的ip以及db
	host := config.GetString("redis.host")
	db := config.GetInt("redis.db")

	password := config.MustString("redis.password", "")

	// 建立连接池
	return &redis.Pool{
		// 从配置文件获取maxidle以及maxactive，取不到则用后面的默认值
		MaxIdle:     config.MustInt("redis.maxidle", 1),
		MaxActive:   config.MustInt("redis.maxactive", 10),
		IdleTimeout: 180 * time.Second,
		Dial: func() (redis.Conn, error) {
			//
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}

			if len(password) > 0 {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()

					logger.Fatal("ooo - redis auth fail, host:\"%s\", password: \"%s\"", host, password)

					return nil, err
				}
			}

			// 选择db
			if db > 0 {
				c.Do("SELECT", db)
			}

			return c, nil
		},
	}
}