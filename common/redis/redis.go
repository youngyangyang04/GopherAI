package redis

import (
	"GopherAI/config"
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

var Rdb *redis.Client

var ctx = context.Background()

func Init() error {
	conf := config.GetConfig()
	host := conf.RedisConfig.RedisHost
	port := conf.RedisConfig.RedisPort
	password := conf.RedisConfig.RedisPassword
	db := conf.RedisDb
	addr := host + ":" + strconv.Itoa(port)

	Rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	if err := Rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	return nil
}

func SetCaptchaForEmail(email, captcha string) error {
	key := GenerateCaptcha(email)
	expire := 2 * time.Minute
	return Rdb.Set(ctx, key, captcha, expire).Err()
}

func CheckCaptchaForEmail(email, userInput string) (bool, error) {
	key := GenerateCaptcha(email)

	// 在 Redis 中获取存储的验证码
	storedCaptcha, err := Rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {

			return false, nil
		}
		return false, err
	}

	if strings.EqualFold(storedCaptcha, userInput) {
		// 验证成功后删除 key
		if err := Rdb.Del(ctx, key).Err(); err != nil {
			// 删除失败，但验证码已验证成功，记录日志即可
			log.Println("Del captcha key failed:", err)
		} else {
			log.Println("Del captcha key success:", key)
		}
		return true, nil
	}

	return false, nil
}
