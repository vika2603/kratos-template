package main

import (
	"gorm.io/gen"

	"kratos-template/pkg/model"
)

// targets lists the per-service query packages to generate. Each service keeps
// its own generated package (internal boundary forbids sharing), regenerated
// from the shared models in pkg/model.
var targets = []string{
	"./app/user/internal/data/query",
	"./app/auth/internal/data/query",
}

func main() {
	for _, out := range targets {
		g := gen.NewGenerator(gen.Config{
			OutPath:      out,
			ModelPkgPath: "kratos-template/pkg/model",
			Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
		})

		g.ApplyBasic(model.User{})
		g.Execute()
	}
}
