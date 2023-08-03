package handler

import (
	"ChopinLogoChangerGo/config"
	"ChopinLogoChangerGo/zlibService"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/gookit/slog"
)

type UnpackHandler struct {
	zlibExtractor *zlibService.Extractor
}

func (handler UnpackHandler) Unpack() error {

	logoFile, err := os.ReadFile(config.Config.LogoFile)
	if err != nil {
		return err
	}
	logoFileReader := bytes.NewReader(logoFile)

	mtk_header := make([]byte, 512)
	_, err = logoFileReader.Read(mtk_header)
	if err != nil && err != io.EOF {
		return err
	}
	if string(mtk_header[8:12]) == "logo" {
		slog.Info("Found 'logo' signature at offset 0x08!")
	} else {
		slog.Info("No 'logo' signature found at offset 0x08, continue anyway...")
	}

	picture_count_bytes := make([]byte, 4)
	_, err = logoFileReader.Read(picture_count_bytes)
	if err != nil && err != io.EOF {
		return err
	}
	picture_count := binary.LittleEndian.Uint16(picture_count_bytes)
	slog.Infof("File contains %d pictures!", picture_count)
	slog.Infof("Reading block size (4 bytes)...")

	bloc_size_bytes := make([]byte, 4)
	_, err = logoFileReader.Read(bloc_size_bytes)
	if err != nil && err != io.EOF {
		return err
	}
	bloc_size := binary.LittleEndian.Uint32(bloc_size_bytes)
	slog.Infof("Total block size (8 bytes + map + pictures): %d", bloc_size)

	offset_map_size := picture_count * 4
	slog.Infof("Reading offsets map (%d * %d = %d bytes)...", picture_count, 4, offset_map_size)

	offsets := map[int]int{}
	sizes := map[int]int{}

	for i := 0; i < int(picture_count); i++ {
		tmp := make([]byte, 4)
		_, err = logoFileReader.Read(tmp)
		if err != nil && err != io.EOF {
			return err
		}
		offsets[i] = int(binary.LittleEndian.Uint32(tmp))
	}

	for i := 0; i < int(picture_count)-1; i++ {
		sizes[i] = offsets[i+1] - offsets[i]
	}

	sizes[int(picture_count)-1] = int(bloc_size) - offsets[int(picture_count)-1]

	images_size := 0
	for i := 0; i < int(picture_count); i++ {
		image_z := make([]byte, sizes[i])
		_, err := logoFileReader.Read(image_z)
		if err != nil && err != io.EOF {
			return err
		}
		images_size += sizes[i]
		binImg, err := handler.zlibExtractor.Extract(image_z)
		if err != nil {
			return err
		}
		binImgFile, err := os.Create(fmt.Sprintf("logo.d/Img%d.bin", i))
		if err != nil {
			return err
		}
		_, err = binImgFile.Write(binImg)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewUnpackHandler(ext *zlibService.Extractor) *UnpackHandler {
	return &UnpackHandler{zlibExtractor: ext}
}
