package main

import (
	"kratos-template/pkg/model"

	"gorm.io/gen"
)

// targets lists models per owning service; auth owns no table.
var targets = map[string][]any{
	"./app/user/internal/data/query": {model.User{}},
}

func main() {
	for out, models := range targets {
		g := gen.NewGenerator(gen.Config{
			OutPath:      out,
			ModelPkgPath: "kratos-template/pkg/model",
			Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		g.ApplyBasic(models...)
		g.Execute()
	}
}
