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
	BGColor      string        `json:"bgColor"` // Hex string, e.g. "#ffffff"
	Format       string        `json:"format"`  // "png" or "jpeg"
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

			// If dimensions match target, bypass scaling
			var scaledImg image.Image
			if srcImg.Bounds().Dx() == placed.Width && srcImg.Bounds().Dy() == placed.Height {
				scaledImg = srcImg
			} else {
				dstRect := image.Rect(0, 0, placed.Width, placed.Height)
				dstImg := image.NewRGBA(dstRect)
				// ponytail: Using BiLinear interpolation for high quality and good speed.
				xdraw.BiLinear.Scale(dstImg, dstRect, srcImg, srcImg.Bounds(), xdraw.Over, nil)
				scaledImg = dstImg
			}

			results <- scaledResult{
				index: index,
				img:   scaledImg,
				x:     placed.X,
				y:     placed.Y,
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
		// Draw res.img at res.x, res.y relative to canvas
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
