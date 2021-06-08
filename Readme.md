# Temp Secure Notes

## Env Vars

 * `PORT`: HTTP port (8080)
 * `MODE`: (debug|release) [default:debug]
 * `LOGGING_LEVEL`: (0:Debug|50:Warning|100:Error) [default:0]
 * `DATA_TTL`: ttl in seconds [default:3600]
 * `REDIS_URL`: (redis://:@localhost:6379)
 * `REDIS_MAX_OPEN`: [default:5]
 * `REDIS_MAX_IDLE`: [default:5]
 * `IN_MEMORY` (false: use redis | true: uses an internal cache as storage ) [default:false]
 * `MAX_NOTE_SIZE` max size in bytes for each note [default:131072 ~ 13kB]

## Debug from VS Code

1. Copy `dev.env_example` to `dev.env` and set redis correct url.

## Storage Methods

### In memory

In memory storage uses [bigcache](https://github.com/allegro/bigcache).

To use this mechanism the env var `IN_MEMORY` must be set to `true`.

### Redis

Redis is used as storage by default. `IN_MEMORY=false`.

To use this mechanism the env var `REDIS_URL` must be set to the redis url server example: `redis://:password@localhost:6379`.
