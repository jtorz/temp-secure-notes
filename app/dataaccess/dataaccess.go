package dataaccess

import (
	"context"
	"log"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gomodule/redigo/redis"
	"github.com/jtorz/temp-secure-notes/app/config"
	"github.com/jtorz/temp-secure-notes/app/ctxinfo"
)

// DataAccces Data Access Interface
type DataAccces interface {
	GetNote(ctx context.Context, key string) (data []byte, found bool, _ error)
	GetVersion(ctx context.Context, key string) (version string, err error)
	SetNote(ctx context.Context, key string, data []byte) (version string, err error)
}

type Redis struct {
	pool *redis.Pool
	ttl  int
}

func NewRedis(pool *redis.Pool, ttl int) *Redis {
	return &Redis{
		pool: pool,
		ttl:  ttl,
	}
}

func (r Redis) GetNote(ctx context.Context, key string) (data []byte, found bool, _ error) {
	con := r.pool.Get()
	defer con.Close()
	data, err := redis.Bytes(con.Do("HGET", key, "NOTE"))
	if err != nil {
		if err == redis.ErrNil {
			return nil, false, nil
		}
		return nil, false, err
	}

	if ctxinfo.LogginAllowed(ctx, config.LogDebug) {
		log.Printf("readed %s", humanize.Bytes(uint64(len(data))))
	}

	return data, true, nil
}

func (r Redis) GetVersion(ctx context.Context, key string) (version string, err error) {
	con := r.pool.Get()
	defer con.Close()
	data, err := redis.String(con.Do("HGET", key, "VERS"))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil
		}
		return "", err
	}
	return data, nil
}

func (r Redis) SetNote(ctx context.Context, key string, data []byte) (version string, err error) {
	con := r.pool.Get()
	defer con.Close()

	if ctxinfo.LogginAllowed(ctx, config.LogDebug) {
		log.Printf("writing %s", humanize.Bytes(uint64(len(data))))
	}

	t := time.Now().Format(time.RFC3339Nano)
	_, err = con.Do("HSET", key, "NOTE", data)
	if err != nil {
		return "", err
	}
	_, err = con.Do("HSET", key, "VERS", t)
	if err != nil {
		return "", err
	}
	_, err = con.Do("EXPIRE", key, r.ttl)
	if err != nil {
		return "", err
	}
	return t, nil
}
