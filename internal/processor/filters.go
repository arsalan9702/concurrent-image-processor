package processor

import (
	"image"
	"image/color"
	"math"

	"github.com/arsalan9702/concurrent-image-processor/internal/models"
)

// Filter represents s function that can be applied to pixel data
type Filter func(src []uint8, width int, params models.FilterParams) []uint8

var FilterRegistry = map[models.FilterType]Filter{
	models.FilterBlur:       ApplyBlur,
	models.FilterBrightness: ApplyBrightness,
	models.FilterConstrast:  ApplyContrast,
	models.FilterGrayScale:  ApplyGrayScale,
}

func ApplyGrayScale(src []uint8, width int, params models.FilterParams) []uint8 {
	if len(src)%4 != 0 {
		return src
	}

	dst := make([]uint8, len(src))

	for i := 0; i < len(src); i += 4 {
		r := float64(src[i])
		g := float64(src[i+1])
		b := float64(src[i+2])
		a := src[i+3]

		gray := uint8(0.299*r + 0.587*g + 0.114*b)

		dst[i] = gray
		dst[i+1] = gray
		dst[i+2] = gray
		dst[i+3] = a
	}

	return dst
}

func ApplyBrightness(src []uint8, width int, params models.FilterParams) []uint8 {
	if len(src)%4 != 0 {
		return src
	}

	dst := make([]uint8, len(src))
	factor := params.Brightness

	for i := 0; i < len(src); i += 4 {
		r := clamp(float64(src[i]) * factor)
		g := clamp(float64(src[i+1]) * factor)
		b := clamp(float64(src[i+2]) * factor)
		a := src[i+3]

		dst[i] = uint8(r)
		dst[i+1] = uint8(g)
		dst[i+2] = uint8(b)
		dst[i+3] = a
	}

	return dst
}

func ApplyContrast(src []uint8, width int, params models.FilterParams) []uint8 {
	if len(src)%4 != 0 {
		return src
	}

	dst := make([]uint8, len(src))
	factor := params.Contrast

	for i := 0; i < len(src); i += 4 {
		r := clamp((float64(src[i]-128) * factor) + 128)
		g := clamp((float64(src[i+1]-128) * factor) + 128)
		b := clamp((float64(src[i+2]-128) * factor) + 128)
		a := src[i+3]

		dst[i] = uint8(r)
		dst[i+1] = uint8(g)
		dst[i+2] = uint8(b)
		dst[i+3] = a
	}

	return dst
}

func ApplyBlur(src []uint8, width int, params models.FilterParams) []uint8 {
	if len(src)%4 != 0 {
		return src
	}

	height := len(src) / (width * 4)
	if height <= 0 {
		return src
	}

	dst := make([]uint8, len(src))
	radius := int(params.BlurRadius)

	if radius <= 0 {
		copy(dst, src)
		return dst
	}

	// Simple box blur implementation
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a float64
			count := 0

			// Sample pixels in the blur radius
			for dy := -radius; dy <= radius; dy++ {
				for dx := -radius; dx <= radius; dx++ {
					nx, ny := x+dx, y+dy
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						idx := (ny*width + nx) * 4
						r += float64(src[idx])
						g += float64(src[idx+1])
						b += float64(src[idx+2])
						a += float64(src[idx+3])
						count++
					}
				}
			}

			if count > 0 {
				idx := (y*width + x) * 4
				dst[idx] = uint8(r / float64(count))
				dst[idx+1] = uint8(g / float64(count))
				dst[idx+2] = uint8(b / float64(count))
				dst[idx+3] = uint8(a / float64(count))
			}
		}
	}

	return dst
}

func ImageToRGBA(img image.Image) *image.RGBA{
	bounds:=img.Bounds()
	rgba:=image.NewRGBA(bounds)

	for y:=bounds.Min.Y; y<bounds.Max.Y; y++{
		for x:=bounds.Min.X; x<bounds.Max.X; x++{
			rgba.Set(x, y, img.At(x, y))
		}
	}

	return rgba
}

func ExtractRowPixels(img *image.RGBA, row int) []uint8 {
	bounds:= img.Bounds()
	widht:=bounds.Dx()

	if row<0 || row>=bounds.Dy(){
		return nil
	}

	pixels:=make([]uint8 , widht*4)
	y:=bounds.Min.Y + row

	for x:=0; x<widht; x++{
		c:=img.RGBAAt(bounds.Min.X+x, y)
		idx:=x*4
		pixels[idx]=c.R
		pixels[idx+1]=c.G
		pixels[idx+2]=c.B
		pixels[idx+3]=c.A
	}

	return pixels
}

func SetRowPixels(img *image.RGBA, row int, pixels []uint8){
	bounds:=img.Bounds()
	width:=bounds.Dx()

	if row<0 || row>=bounds.Dy() || len(pixels)!=width*	4{
		return
	}

	y:=bounds.Min.Y+row

	for x := 0; x < width; x++ {
		idx := x * 4
		c := color.RGBA{
			R: pixels[idx],
			G: pixels[idx+1],
			B: pixels[idx+2],
			A: pixels[idx+3],
		}
		img.SetRGBA(bounds.Min.X+x, y, c)
	}
}

// clamp ensures value is within 0-255 range
func clamp(value float64) float64 {
	return math.Max(0, math.Min(255, value))
}
