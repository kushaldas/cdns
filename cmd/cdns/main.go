package main

import (
	"github.com/kushaldas/cdns/pkg/mserver"
	flag "github.com/spf13/pflag"
)

var port *int = flag.Int("port", 53, "The port to run on.")

var remote *string = flag.String("remote", "1.1.1.1:53", "The remote DNS IP:PORT to be used.")

func main() {

	flag.Parse()

	mserver.Listen(*port, *remote)
}
