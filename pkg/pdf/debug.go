package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	"github.com/signintech/pdft"
)

// Debug overlays the frame boundaries on a PDF page for visual inspection.
//
// It draws two rectangles:
//   - the outer frame (red), which corresponds to the raw X/Y/W/H of the frame;
//   - the inner frame (blue), which corresponds to the text area after applying margins.
//
// This method is intended for debugging and layout tuning only.
// It helps to verify that text and images are aligned properly within the frame.
func (f Frame) Debug(pdf *pdft.PDFt, page int) error {
	// generate and draw outer frame
	err := f.drawRect(
		pdf, page,
		f.X, f.Y, f.W, f.H, // outer frame's coordinates
		color.RGBA{R: 255, G: 0, B: 0, A: 255})
	if err != nil {
		return err
	}

	// generate and draw inner frame (using margin)
	ix, iy, iw, ih := f.InnerRect() // inner frame's coordinates
	err = f.drawRect(
		pdf, page,
		ix, iy, iw, ih,
		color.RGBA{R: 0, G: 0, B: 255, A: 255})
	if err != nil {
		return err
	}

	return nil
}

// drawRect inserts a colored rectangular outline as a PNG image into the PDF.
//
// The rectangle is rendered as a transparent image with a 1-pixel border of the given color,
// and placed at the specified coordinates (x, y) with the specified width and height (w, h).
// The (x, y) coordinates correspond to the top-left corner of the rectangle.
func (f Frame) drawRect(pdf *pdft.PDFt, page int, x, y, w, h float64, color color.RGBA) error {
	img, err := generateFrameImage(int(w), int(h), color)
	if err != nil {
		return err
	}
	return pdf.InsertImg(img, page, x, y, w, h)
}

// generateFrameImage returns a PNG image of a transparent rectangle with a colored border.
// The border has a fixed thickness of 1 px.
func generateFrameImage(w, h int, frameColor color.RGBA) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// полностью прозрачный фон
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{}}, image.Point{}, draw.Src)

	// (толщина = 1px)
	for x := 0; x < w; x++ {
		img.Set(x, 0, frameColor)
		img.Set(x, h-1, frameColor)
	}
	for y := 0; y < h; y++ {
		img.Set(0, y, frameColor)
		img.Set(w-1, y, frameColor)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
