package main

import (
	"flag"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	var cfg Config
	var flags flag.FlagSet
	flags.StringVar(&cfg.Title, "title", "API", "API title")
	flags.StringVar(&cfg.Version, "version", "0.1.0", "API version")
	flags.StringVar(&cfg.Description, "description", "", "API description")
	flags.StringVar(&cfg.Output, "output", "openapi.yaml", "output file path")
	flags.StringVar(&cfg.Naming, "naming", "json", "field naming: json (camelCase) or proto (snake_case)")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		return Generate(gen, &cfg)
	})
}
