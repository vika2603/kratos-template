package main

import (
	"gorm.io/gen"

	"kratos-template/pkg/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:      "./app/user/internal/data/query",
		ModelPkgPath: "kratos-template/pkg/model",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.ApplyBasic(model.User{})
	g.Execute()

	g2 := gen.NewGenerator(gen.Config{
		OutPath:      "./app/asset/internal/data/query",
		ModelPkgPath: "kratos-template/pkg/model",
		Mode:         gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g2.ApplyBasic(model.Asset{})
	g2.Execute()
}
