# Gossip Glomers – Go

This repository contains my attempts at the distributed systems challenge
"[Gossip Glomers](https://fly.io/dist-sys/)" by the folks at fly.io

## Running an application

I typically run an application via a one-liner.

This one-liner does two things.

1. Compiles the binary, outputting it to the `bin/` directory.
2. Runs the output binary with `maelstrom`.

```zsh
go build -o bin/echo.bin ./cmd/01/main.go && maelstrom test -w echo --bin bin/echo.bin --node-count 1 --time-limit 10
```