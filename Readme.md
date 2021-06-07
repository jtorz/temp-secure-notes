# Temp Secure Notes

## Requirements

* Redis

## Env Vars

 * `PORT`: HTTP port (8080)
 * `MODE`: (debug|release) [default:debug]
 * `LOGGING_LEVEL`: (0:Debug|50:Warning|100:Error) [default:0]
 * `DATA_TTL`: ttl in seconds [default:3600]
 * `REDIS_URL`: (redis://:@localhost:6379)
 * `REDIS_MAX_OPEN`: [default:5]
 * `REDIS_MAX_IDLE`: [default:5]

## Debug from VS Code

1. Copy `dev.env_example` to `dev.env` and set redis correct url.