package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/pilu/go-base62"
)

const (
	URLIdKey           = "next.url.id"         // global counter
	ShortLinkKey       = "shortlink:%s:url"    // mapping the short link to the original url
	URLHashKey         = "urlhash:%s:url"      // mapping the url hash to the original url
	ShortLinkDetailKey = "shortlink:%s:detail" // mapping the short link to the detail of url
)

type RedisClient struct {
	cli *redis.Client
}

type ShortLinkInfo struct {
	URL                 string        `json:"url"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
	CreatedAt           string        `json:"created_at"`
}

func NewRedisClient(rc *RedisConf) *RedisClient {
	addr := fmt.Sprintf("%s:%d", rc.Host, rc.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: rc.Password,
		DB:       rc.DB,
	})
	if _, err := client.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisClient{client}
}

func (r *RedisClient) Shorten(url string, expiration int64) (string, error) {
	urlHash := toSHA1(url)

	d, err := r.cli.Get(fmt.Sprintf(URLHashKey, urlHash)).Result()
	if err == redis.Nil {
		// not existed, nothing to do
	} else if err != nil {
		return "", err
	} else {
		if d == "{}" {
			// expiration, nothing to do
		} else {
			return d, nil
		}
	}

	// 1. increase the global counter
	id, err := r.cli.Incr(URLIdKey).Result()
	if err != nil {
		return "", err
	}

	// encode global counter to base62
	encodeId := base62.Encode(int(id))

	// 2. save shortlink to redis
	err = r.cli.Set(fmt.Sprintf(ShortLinkKey, encodeId), url, time.Minute*time.Duration(expiration)).Err()
	if err != nil {
		return "", err
	}

	// 3. save hash value and short link
	err = r.cli.Set(fmt.Sprintf(URLHashKey, urlHash), encodeId, time.Minute*time.Duration(expiration)).Err()
	if err != nil {
		return "", err
	}

	info, err := json.Marshal(&ShortLinkInfo{
		URL:                 url,
		ExpirationInMinutes: time.Duration(expiration),
		CreatedAt:           time.Now().String(),
	})
	if err != nil {
		return "", err
	}

	// 4. save short link and link detail
	err = r.cli.Set(fmt.Sprintf(ShortLinkDetailKey, encodeId), info, time.Minute*time.Duration(expiration)).Err()
	if err != nil {
		return "", nil
	}

	return encodeId, nil
}

func (r *RedisClient) UnShorten(encodeId string) (string, error) {
	url, err := r.cli.Get(fmt.Sprintf(ShortLinkKey, encodeId)).Result()
	if err == redis.Nil {
		return "", StatusError{
			Code: 404,
			Err:  errors.New("short link not found"),
		}
	} else if err != nil {
		return "", err
	}

	return url, nil
}

func (r *RedisClient) ShortLinkInfo(encodeId string) (ShortLinkInfo, error) {
	var info ShortLinkInfo
	detail, err := r.cli.Get(fmt.Sprintf(ShortLinkDetailKey, encodeId)).Result()
	if err == redis.Nil {
		return info, StatusError{
			Code: 404,
			Err:  errors.New("short link not found"),
		}
	} else if err != nil {
		return info, err
	}

	err = json.Unmarshal([]byte(detail), &info)
	if err != nil {
		return info, err
	}
	return info, nil

}

func toSHA1(data string) string {
	h := sha1.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
