package main

import (
	"github.com/igvaquero18/packer_doconsul/doconsul"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterPostProcessor(new(doconsul.PostProcessor))
	server.Serve()
}
