package main

import (
	"gorm.io/gen"

	"kratos-template/pkg/model"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./app/user/internal/data/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	g.ApplyBasic(model.User{})

	g.Execute()
}
