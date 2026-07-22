package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"

	"golang.org/x/image/draw"
)

var (
	profileImg  image.Image
	projectImgs = map[string]image.Image{}
)

func loadImages(assetsDir string) {
	profileImg, _ = decodeImage(assetsDir + "/about/egor.jpeg")
	for _, pair := range [][2]string{
		{"grape", assetsDir + "/projects/grape.jpeg"},
		{"federated", assetsDir + "/projects/federated.jpeg"},
		{"nix-ml-solo", assetsDir + "/projects/nix-ml-solo.jpeg"},
		{"nixlab", assetsDir + "/projects/nixlab.jpeg"},
	} {
		if img, err := decodeImage(pair[1]); err == nil {
			projectImgs[pair[0]] = img
		}
	}
}

func decodeImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	return img, err
}

func scaleImg(src image.Image, w, h int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// halfBlock renders img using half-block (▄) characters.
// targetW is in terminal columns. Returns "" if img is nil.
// Each terminal char cell is treated as 2 pseudo-pixels tall and 1 wide,
// so the rendered image maintains correct aspect ratio on typical 2:1 fonts.
func halfBlock(img image.Image, targetW int) string {
	if img == nil || targetW <= 0 {
		return ""
	}
	bounds := img.Bounds()
	srcW := bounds.Max.X - bounds.Min.X
	srcH := bounds.Max.Y - bounds.Min.Y
	if srcW == 0 || srcH == 0 {
		return ""
	}

	pixH := targetW * srcH / srcW
	rows := (pixH + 1) / 2
	if rows == 0 {
		rows = 1
	}

	scaled := scaleImg(img, targetW, rows*2)

	var sb strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < targetW; col++ {
			top := scaled.NRGBAAt(col, row*2)
			bot := scaled.NRGBAAt(col, row*2+1)
			sb.WriteString(fmt.Sprintf(
				"\033[48;2;%d;%d;%dm\033[38;2;%d;%d;%dm▄",
				top.R, top.G, top.B,
				bot.R, bot.G, bot.B,
			))
		}
		sb.WriteString("\033[0m\n")
	}
	return sb.String()
}

// centeredHalfBlock renders img centered within totalW columns.
func centeredHalfBlock(img image.Image, targetW, totalW int) string {
	raw := halfBlock(img, targetW)
	if raw == "" {
		return ""
	}
	pad := strings.Repeat(" ", max(0, (totalW-targetW)/2))
	var sb strings.Builder
	for _, line := range strings.Split(strings.TrimRight(raw, "\n"), "\n") {
		sb.WriteString(pad + line + "\n")
	}
	return sb.String()
}
