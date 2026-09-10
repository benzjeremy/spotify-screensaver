package palette

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"net/http"
	"sort"
	"sync"
	"time"
)

// ExtractedColors contains the primary and secondary hex colors.
type ExtractedColors struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// Extractor extracts dominant colors from images with caching.
type Extractor struct {
	mu         sync.RWMutex
	cache      map[string]ExtractedColors
	httpClient *http.Client
}

// NewExtractor creates a new Extractor instance.
func NewExtractor() *Extractor {
	return &Extractor{
		cache: make(map[string]ExtractedColors),
		httpClient: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

// ExtractFromURL fetches an image by URL and extracts its dominant colors.
func (e *Extractor) ExtractFromURL(url string) (ExtractedColors, error) {
	if url == "" {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, nil
	}

	e.mu.RLock()
	if colors, found := e.cache[url]; found {
		e.mu.RUnlock()
		return colors, nil
	}
	e.mu.RUnlock()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, err
	}
	req.Header.Set("User-Agent", "SpotifyScreensaver/1.3")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, fmt.Errorf("http error: %d", resp.StatusCode)
	}

	colors, err := e.ExtractFromReader(resp.Body)
	if err != nil {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, err
	}

	e.mu.Lock()
	e.cache[url] = colors
	e.mu.Unlock()

	return colors, nil
}

// ExtractFromReader decodes an image stream and extracts dominant colors.
func (e *Extractor) ExtractFromReader(r io.Reader) (ExtractedColors, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}, err
	}
	return ExtractDominantColors(img), nil
}

type colorCount struct {
	r, g, b uint32
	count   int
	weight  float64
}

// ExtractDominantColors extracts primary and secondary colors using quantizing & vibrance scoring.
func ExtractDominantColors(img image.Image) ExtractedColors {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}
	}

	// Sample pixels with a stride for high performance (<5ms)
	strideX := int(math.Max(1, float64(width)/48.0))
	strideY := int(math.Max(1, float64(height)/48.0))

	buckets := make(map[uint32]*colorCount)

	for y := bounds.Min.Y; y < bounds.Max.Y; y += strideY {
		for x := bounds.Min.X; x < bounds.Max.X; x += strideX {
			c := img.At(x, y)
			r, g, b, a := c.RGBA()
			if a < 32768 {
				continue // Skip mostly transparent pixels
			}
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Quantize 8-bit to 4-bit (16 levels per channel = 4096 bins)
			qr := (r8 >> 4) << 4
			qg := (g8 >> 4) << 4
			qb := (b8 >> 4) << 4
			key := (uint32(qr) << 16) | (uint32(qg) << 8) | uint32(qb)

			entry, exists := buckets[key]
			if !exists {
				entry = &colorCount{
					r: uint32(qr),
					g: uint32(qg),
					b: uint32(qb),
				}
				buckets[key] = entry
			}
			entry.count++
		}
	}

	if len(buckets) == 0 {
		return ExtractedColors{Primary: "#1db954", Secondary: "#121212"}
	}

	// Rank buckets by vibrance and count
	var candidates []*colorCount
	for _, b := range buckets {
		rf := float64(b.r) / 255.0
		gf := float64(b.g) / 255.0
		bf := float64(b.b) / 255.0

		max := math.Max(rf, math.Max(gf, bf))
		min := math.Min(rf, math.Min(gf, bf))
		delta := max - min
		saturation := 0.0
		if max > 0.001 {
			saturation = delta / max
		}
		brightness := (max + min) / 2.0

		// Ignore near-black (<0.08) and near-white (>0.94) to favor vibrant colors
		if brightness < 0.08 || brightness > 0.94 {
			b.weight = float64(b.count) * 0.1
		} else {
			// Favor vibrant saturation
			b.weight = float64(b.count) * (0.4 + saturation*1.8)
		}
		candidates = append(candidates, b)
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].weight > candidates[j].weight
	})

	primary := candidates[0]
	primaryHex := fmt.Sprintf("#%02x%02x%02x", primary.r, primary.g, primary.b)

	// Find distinct secondary color
	var secondary *colorCount
	for i := 1; i < len(candidates); i++ {
		cand := candidates[i]
		dist := colorDistance(primary, cand)
		if dist > 60.0 {
			secondary = cand
			break
		}
	}

	if secondary == nil {
		// Complementary / darker offset
		secR := (primary.r * 2) / 3
		secG := (primary.g * 2) / 3
		secB := (primary.b * 2) / 3
		return ExtractedColors{
			Primary:   primaryHex,
			Secondary: fmt.Sprintf("#%02x%02x%02x", secR, secG, secB),
		}
	}

	secondaryHex := fmt.Sprintf("#%02x%02x%02x", secondary.r, secondary.g, secondary.b)
	return ExtractedColors{
		Primary:   primaryHex,
		Secondary: secondaryHex,
	}
}

func colorDistance(c1, c2 *colorCount) float64 {
	dr := float64(c1.r) - float64(c2.r)
	dg := float64(c1.g) - float64(c2.g)
	db := float64(c1.b) - float64(c2.b)
	return math.Sqrt(dr*dr + dg*dg + db*db)
}

// ConvertColorToHex converts a standard color to hex format.
func ConvertColorToHex(c color.Color) string {
	r, g, b, _ := c.RGBA()
	return fmt.Sprintf("#%02x%02x%02x", uint8(r>>8), uint8(g>>8), uint8(b>>8))
}
