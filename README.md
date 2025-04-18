# genericpubsub

[![Go Tests](https://github.com/sesopenko/genericpubsub/actions/workflows/test.yml/badge.svg)](https://github.com/sesopenko/genericpubsub/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sesopenko/genericpubsub.svg)](https://pkg.go.dev/github.com/sesopenko/genericpubsub)
[![GitHub tag](https://img.shields.io/github/v/tag/sesopenko/genericpubsub?label=version)](https://github.com/sesopenko/genericpubsub/tags)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.txt)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue)](https://golang.org/doc/go1.24)


`genericpubsub` is a lightweight, type-safe publish/subscribe system for Go, using generics and context-based
cancellation. It allows you to publish messages of any type to multiple subscribers concurrently, with clean shutdown
and resource management.

## Features

- Type-safe using Go generics
- Simple API: `Publish`, `Subscribe`
- Context-based cancellation
- Graceful shutdown of subscribers
- Fully tested with unit tests

## Installation

```bash
go get github.com/sesopenko/genericpubsub
````

## Documentation

https://pkg.go.dev/github.com/sesopenko/genericpubsub

## Example

```go
package main

import (
    "context"
    "fmt"
	"time"
    "github.com/sesopenko/genericpubsub"
)

type Message struct {
    Value string
}

func main() {
    channelBuffer := 64
    ps := genericpubsub.New[Message](context.Background(), channelBuffer)
    sub := ps.Subscribe(context.TODO(), channelBuffer)
    
    go ps.Send(Message{Value: "hello"})
    time.Sleep(50 * time.Millisecond)
    msg, ok := <-sub
    fmt.Println("Received:", msg.Value)
    fmt.Printf("channel wasn't closed: %t\n", ok)
}

```

## License

This project is licensed under the MIT license. See [LICENSE.txt](LICENSE.txt) for details.

© 2025 Sean Esopenko