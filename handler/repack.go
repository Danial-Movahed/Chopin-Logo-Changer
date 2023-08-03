package handler

import "ChopinLogoChangerGo/zlibService"

type RepackHandler struct {
	zlibCompressor *zlibService.Compressor
}

func (handler RepackHandler) Repack() {

}

func NewRepackHandler(cmp *zlibService.Compressor) *RepackHandler {
	return &RepackHandler{zlibCompressor: cmp}
}
