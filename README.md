# Hot cache

## Build

GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o ./hot_cache_linux -a
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -o ./hot_cache_macos -a
