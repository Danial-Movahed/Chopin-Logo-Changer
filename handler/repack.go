package handler

import (
	"ChopinLogoChangerGo/config"
	"ChopinLogoChangerGo/zlibService"
	"encoding/binary"
	"os"
	"sort"
	"strings"

	"github.com/gookit/slog"
)

type RepackHandler struct {
	zlibCompressor *zlibService.Compressor
}

func (handler RepackHandler) copyFile(src string, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (handler RepackHandler) Repack() error {
	handler.copyFile(config.Config.LogoFile, config.Config.OutputFile)
	outputFile, err := os.OpenFile(config.Config.OutputFile, os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	files, err := os.ReadDir("logo.d")
	if err != nil {
		return err
	}
	var images []string
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "Img") && strings.HasSuffix(file.Name(), ".bin") {
			images = append(images, file.Name())
		}
	}
	sort.Strings(images)

	picture_count := len(images)
	bloc_size := (4 * 2) + (4 * picture_count)
	pos := int64(512)
	offsets := map[int]int{}
	sizes := map[int]int{}
	imgsCompressed := [][]byte{}
	picture_count_byte := make([]byte, 4)
	binary.LittleEndian.PutUint32(picture_count_byte, uint32(picture_count))
	_, err = outputFile.WriteAt(picture_count_byte, pos)
	if err != nil {
		return err
	}
	pos += 4
	i := 0
	for _, img := range images {
		imgBytes, err := os.ReadFile("logo.d/" + img)
		if err != nil {
			return err
		}
		imgCompressed, err := handler.zlibCompressor.Compress(imgBytes)
		if err != nil {
			return err
		}
		imgsCompressed = append(imgsCompressed, imgCompressed)
		sizes[i] = len(imgCompressed)
		if i == 0 {
			offsets[i] = bloc_size
		} else {
			offsets[i] = offsets[i-1] + sizes[i-1]
		}
		bloc_size += sizes[i]
		i += 1
	}

	slog.Infof("Total block size (8 bytes + map + pictures): %d", bloc_size)
	slog.Infof("Writing total block size (%d bytes)...", 4)
	bloc_size_bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bloc_size_bytes, uint32(bloc_size))
	_, err = outputFile.WriteAt(bloc_size_bytes, pos)
	if err != nil {
		return err
	}
	pos += 4
	slog.Infof("Writing offsets map (%d * %d = %d bytes)", picture_count, 4, picture_count*4)
	images_size := 0
	for i := 0; i < picture_count; i++ {
		offset_bytes := make([]byte, 4)
		binary.LittleEndian.PutUint32(offset_bytes, uint32(offsets[i]))
		_, err = outputFile.WriteAt(offset_bytes, pos)
		if err != nil {
			return err
		}
		pos += 4
		images_size += sizes[i]
	}

	for _, imgCmp := range imgsCompressed {
		outputFile.WriteAt(imgCmp, pos)
		pos += int64(len(imgCmp))
	}

	return nil
}

func NewRepackHandler(cmp *zlibService.Compressor) *RepackHandler {
	return &RepackHandler{zlibCompressor: cmp}
}
