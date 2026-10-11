// Command server runs the service: on start it applies the database migrations, then serves
// Connect, gRPC, gRPC-Web and REST on APP_ADDR.
package main

import "github.com/tanjed/base-go/internal/ioc"

func main() {
	ioc.New().Run()
}
