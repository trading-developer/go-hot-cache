# Hot cache

## Build

GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o ./hot_cache_macos -a

## Compile in docker

```
1. docker-compose build
2. docker-compose up
```