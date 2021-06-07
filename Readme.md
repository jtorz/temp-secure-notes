# Temp Secure Notes

## Requirements

* Redis

## Env Vars

 * `PORT`: HTTP port (8080)
 * `MODE`: (debug|release)
 * `LOGGING_LEVEL`: (0-255)
 * `REDIS_URL`: (redis://:@localhost:6379)
 * `REDIS_MAX_OPEN`: (5)
 * `REDIS_MAX_IDLE`: (5)

## Debug from VS Code

1. Copy `dev.env_example` to `dev.env` and set redis correct url.