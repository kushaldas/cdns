package main

import (
	"github.com/kushaldas/cdns/pkg/mserver"
	flag "github.com/spf13/pflag"
)

var port *int = flag.Int("port", 53, "The port to run on.")

func main() {

	flag.Parse()

	mserver.Listen(*port, "127.0.0.53:53")
}
