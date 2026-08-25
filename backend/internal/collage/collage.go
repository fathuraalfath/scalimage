package collage

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"strconv"
	"strings"
	"sync"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"

	"scalimage/internal/storage"
)

// PlacedImage describes an image placed on the collage canvas.
type PlacedImage struct {
	ID     string `json:"id"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// CollageRequest describes the parameters for generating a collage.
type CollageRequest struct {
	CanvasWidth  int           `json:"canvasWidth"`
	CanvasHeight int           `json:"canvasHeight"`
	BGColor      string        `json:"bgColor"`      // Hex string, e.g. "#ffffff"
	Format       string        `json:"format"`       // "png" or "jpeg"
	Gap          int           `json:"gap"`          // Gap spacing in pixels between images
	BorderRadius int           `json:"borderRadius"` // Corner rounding radius in pixels
	Images       []PlacedImage `json:"images"`
}

// Generator manages the creation of collages.
type Generator struct {
	store storage.Storage
}

// NewGenerator creates a new Generator.
func NewGenerator(store storage.Storage) *Generator {
	return &Generator{store: store}
}

// Generate combines the placed images onto a single canvas.
func (g *Generator) Generate(ctx context.Context, req CollageRequest, out io.Writer) error {
	if req.CanvasWidth <= 0 || req.CanvasHeight <= 0 {
		return fmt.Errorf("invalid canvas dimensions: %dx%d", req.CanvasWidth, req.CanvasHeight)
	}

	// 1. Create target canvas image
	canvas := image.NewRGBA(image.Rect(0, 0, req.CanvasWidth, req.CanvasHeight))

	// 2. Parse and fill background color
	bgColor, err := ParseHexColor(req.BGColor)
	if err != nil {
		// ponytail: default to white on parsing error
		bgColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

	// 3. Concurrent decoding and resizing
	type scaledResult struct {
		index int
		img   image.Image
		x, y  int
		err   error
	}

	results := make(chan scaledResult, len(req.Images))
	var wg sync.WaitGroup

	for idx, pi := range req.Images {
		wg.Add(1)
		go func(index int, placed PlacedImage) {
			defer wg.Done()

			rc, err := g.store.Get(ctx, placed.ID)
			if err != nil {
				results <- scaledResult{index: index, err: fmt.Errorf("failed to read file %q: %w", placed.ID, err)}
				return
			}
			defer rc.Close()

			srcImg, _, err := image.Decode(rc)
			if err != nil {
				results <- scaledResult{index: index, err: fmt.Errorf("failed to decode image %q: %w", placed.ID, err)}
				return
			}

			// Apply gap inset calculations
			targetX := placed.X
			targetY := placed.Y
			targetW := placed.Width
			targetH := placed.Height

			if req.Gap > 0 {
				halfGap := req.Gap / 2
				targetX += halfGap
				targetY += halfGap
				targetW -= req.Gap
				targetH -= req.Gap
				if targetW <= 0 {
					targetW = 1
				}
				if targetH <= 0 {
					targetH = 1
				}
			}

			dstRect := image.Rect(0, 0, targetW, targetH)
			dstImg := image.NewRGBA(dstRect)
			// ponytail: Using BiLinear interpolation for high quality and good speed.
			xdraw.BiLinear.Scale(dstImg, dstRect, srcImg, srcImg.Bounds(), xdraw.Over, nil)

			// Apply corner radius if requested
			if req.BorderRadius > 0 {
				ApplyCornerRadius(dstImg, req.BorderRadius)
			}

			results <- scaledResult{
				index: index,
				img:   dstImg,
				x:     targetX,
				y:     targetY,
			}
		}(idx, pi)
	}

	wg.Wait()
	close(results)

	// 4. Staging the scaled layers
	orderedResults := make([]scaledResult, len(req.Images))
	for res := range results {
		if res.err != nil {
			return res.err
		}
		orderedResults[res.index] = res
	}

	// 5. Sequentially draw layers onto the main canvas (respects layering order)
	for _, res := range orderedResults {
		if res.img == nil {
			continue
		}
		bounds := res.img.Bounds()
		dp := image.Pt(res.x, res.y)
		dr := bounds.Add(dp)
		draw.Draw(canvas, dr, res.img, image.Point{}, draw.Over)
	}

	// 6. Encode output format
	format := strings.ToLower(req.Format)
	if format == "jpeg" || format == "jpg" {
		return jpeg.Encode(out, canvas, &jpeg.Options{Quality: 90})
	}

	// Default to PNG
	return png.Encode(out, canvas)
}

// ApplyCornerRadius masks out pixels outside the specified corner radius.
func ApplyCornerRadius(img *image.RGBA, r int) {
	if r <= 0 {
		return
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if r > w/2 {
		r = w / 2
	}
	if r > h/2 {
		r = h / 2
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var dx, dy float64
			isCorner := false

			if x < r && y < r { // top-left corner
				dx, dy = float64(r-x), float64(r-y)
				isCorner = true
			} else if x >= w-r && y < r { // top-right corner
				dx, dy = float64(x-(w-r-1)), float64(r-y)
				isCorner = true
			} else if x < r && y >= h-r { // bottom-left corner
				dx, dy = float64(r-x), float64(y-(h-r-1))
				isCorner = true
			} else if x >= w-r && y >= h-r { // bottom-right corner
				dx, dy = float64(x-(w-r-1)), float64(y-(h-r-1))
				isCorner = true
			}

			if isCorner {
				dist := math.Hypot(dx, dy)
				if dist > float64(r) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
				}
			}
		}
	}
}

// ParseHexColor parses a hex color string (e.g. "#ffffff" or "fff") into color.RGBA.
func ParseHexColor(s string) (color.RGBA, error) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("invalid hex color string: %s", s)
	}
	r, err := strconv.ParseUint(s[0:2], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	g, err := strconv.ParseUint(s[2:4], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	b, err := strconv.ParseUint(s[4:6], 16, 8)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil
}
