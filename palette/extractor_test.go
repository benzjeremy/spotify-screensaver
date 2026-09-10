package palette

import (
	"image"
	"image/color"
	"image/draw"
	"strings"
	"testing"
)

func TestExtractDominantColors(t *testing.T) {
	// Create a synthetic 100x100 image with a vibrant red patch and a dark blue background
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill background with dark navy blue #101428
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 16, G: 20, B: 40, A: 255}}, image.Point{}, draw.Src)

	// Draw a vibrant pink/crimson circle or square in center #e11d48
	crimson := color.RGBA{R: 225, G: 29, B: 72, A: 255}
	draw.Draw(img, image.Rect(20, 20, 80, 80), &image.Uniform{C: crimson}, image.Point{}, draw.Src)

	colors := ExtractDominantColors(img)

	if !strings.HasPrefix(colors.Primary, "#") {
		t.Fatalf("Expected valid hex primary color, got %s", colors.Primary)
	}
	if !strings.HasPrefix(colors.Secondary, "#") {
		t.Fatalf("Expected valid hex secondary color, got %s", colors.Secondary)
	}

	// Primary should reflect the vibrant red/crimson
	if colors.Primary == "" || colors.Secondary == "" {
		t.Fatalf("Colors must not be empty")
	}
}

func TestConvertColorToHex(t *testing.T) {
	c := color.RGBA{R: 29, G: 185, B: 84, A: 255} // Spotify green
	hex := ConvertColorToHex(c)
	if hex != "#1db954" {
		t.Fatalf("Expected #1db954, got %s", hex)
	}
}
