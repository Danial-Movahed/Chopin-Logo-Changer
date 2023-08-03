package main

import (
	"ChopinLogoChangerGo/config"
	"ChopinLogoChangerGo/handler"
	"ChopinLogoChangerGo/zlibService"
	"flag"
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

func Run(UnpackHndl *handler.UnpackHandler, RepackHndl *handler.RepackHandler) {
	flag.StringVar(&config.Config.LogoFile, "logo", "logo.img", "Original logo file")
	flag.StringVar(&config.Config.OutputFile, "output", "out.bin", "Output logo file")
	flag.Parse()
	if os.Args[1] == "unpack" {
		UnpackHndl.Unpack()
	} else if os.Args[1] == "repack" {
		RepackHndl.Repack()
	}
}
