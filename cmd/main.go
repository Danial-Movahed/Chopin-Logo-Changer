package main

import (
	"ChopinLogoChangerGo/handler"
	"ChopinLogoChangerGo/zlibService"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			zlibService.NewZlibExtractor,
			zlibService.NewZlibCompressor,
			handler.NewRepackHandler,
			handler.NewUnpackHandler,
		),
	)
	app.Run()
}
