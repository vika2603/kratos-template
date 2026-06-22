package main

import (
	"gorm.io/gen"

	"kratos-template/pkg/model"
)

// targets lists the per-service query packages to generate. auth owns no table
// (it reads users via the user service), so only user is here.
var targets = []string{
	"./app/user/internal/data/query",
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
