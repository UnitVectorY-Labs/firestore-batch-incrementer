
# Commands for firestore-batch-incrementer
default:
  @just --list
# Build firestore-batch-incrementer with Go
build:
  go build ./...

# Run tests for firestore-batch-incrementer with Go
test:
  go clean -testcache
  go test ./...