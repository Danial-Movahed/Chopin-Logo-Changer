package main

import (
	"ChopinLogoChangerGo/handler"
	"ChopinLogoChangerGo/zlibService"
	"fmt"
	"os"

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
		fx.Invoke(
			Run,
		),
	)
	app.Run()
}

func Run() {
	fmt.Println(os.Args)
}
