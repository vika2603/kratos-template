package main

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"go.yaml.in/yaml/v4"
	"google.golang.org/protobuf/compiler/protogen"
)

func EmitYAML(doc *v3.Document, g *protogen.GeneratedFile) error {
	enc := yaml.NewEncoder(g)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return enc.Close()
}
