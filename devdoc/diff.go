// +build ignore

package main

import (
	"fmt"
	"log"

	"github.com/golang/protobuf/proto"
	plugin "github.com/golang/protobuf/protoc-gen-go/plugin"
	"github.com/shurcooL/httpfs/vfsutil"
	"src.sourcegraph.com/sourcegraph/devdoc/assets"
)

func main() {
	// Unmarshal the Protobuf-encoded request.
	docs := new(plugin.CodeGeneratorRequest)
	protoRequest, err := vfsutil.ReadFile(assets.Data, "/sourcegraph.dump")
	if err != nil {
		log.Fatalln(err)
	}
	if err := proto.Unmarshal(protoRequest, docs); err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%+v\n", docs)
}
