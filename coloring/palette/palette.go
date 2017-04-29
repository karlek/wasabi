package main

import (
	"fmt"
	"image"
	"math"
	"sort"

	gwc "github.com/jyotiska/go-webcolors"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nfnt/resize"
)

// This method finds the closest color for a given RGB tuple and returns the name of the color in given mode
func FindClosestColor(RequestedColor []int, mode string) string {
	MinColors := make(map[int]string)
	var ColorMap map[string]string

	// css3 gives the shades while css21 gives the primary or base colors
	if mode == "css3" {
		ColorMap = gwc.CSS3NamesToHex
	} else {
		ColorMap = gwc.HTML4NamesToHex
	}

	for name, hexcode := range ColorMap {
		rgb_triplet := gwc.HexToRGB(hexcode)
		rd := math.Pow(float64(rgb_triplet[0]-RequestedColor[0]), float64(2))
		gd := math.Pow(float64(rgb_triplet[1]-RequestedColor[1]), float64(2))
		bd := math.Pow(float64(rgb_triplet[2]-RequestedColor[2]), float64(2))
		MinColors[int(rd+gd+bd)] = name
	}

	keys := make([]int, 0, len(MinColors))
	for key := range MinColors {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return MinColors[keys[0]]
}

// This method creates a reverse map
func ReverseMap(m map[string]int) map[int]string {
	n := make(map[int]string)
	for k, v := range m {
		n[v] = k
	}
	return n
}

// This table contains the "keypoints" of the colorgradient you want to generate.
// The position of each keypoint has to live in the range [0,1]

// This is the meat of the gradient computation. It returns a HCL-blend between
// the two colors around `t`.
// Note: It relies heavily on the fact that the gradient keypoints are sorted.
func (self GradientTable) GetInterpolatedColorFor(t float64) colorful.Color {
	for i := 0; i < len(self)-1; i++ {
		c1 := self[i]
		c2 := self[i+1]
		if c1.Pos <= t && t <= c2.Pos {
			// We are in between c1 and c2. Go blend them!
			t := (t - c1.Pos) / (c2.Pos - c1.Pos)
			return c1.Col.BlendHcl(c2.Col, t).Clamped()
		}
	}

	// Nothing found? Means we're at (or past) the last gradient keypoint.
	return self[len(self)-1].Col
}

type GradientTable []struct {
	Col colorful.Color
	Pos float64
}

func createPalette(img image.Image, limit int) GradientTable {
	// Resize the image to smaller scale for faster computation
	img = resize.Resize(100, 0, img, resize.Bilinear)
	bounds := img.Bounds()

	colorCounter := make(map[string]int)

	for i := 0; i <= bounds.Max.X; i++ {
		for j := 0; j <= bounds.Max.Y; j++ {
			pixel := img.At(i, j)
			red, green, blue, alpha := pixel.RGBA()
			if alpha == 0 {
				continue
			}
			rgbTuple := []int{int(red / 255), int(green / 255), int(blue / 255)}
			colorName := FindClosestColor(rgbTuple, "css3")
			// colorName := hash(pixel.RGBA())
			_, present := colorCounter[colorName]
			if present {
				colorCounter[colorName] += 1
			} else {
				colorCounter[colorName] = 1
			}
		}
	}

	// Sort by the frequency of each color
	keys := make([]int, 0, len(colorCounter))
	for _, val := range colorCounter {
		keys = append(keys, val)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))

	reverseColorCounter := ReverseMap(colorCounter)

	// Display the top N dominant colors from the image
	if len(keys) < limit {
		limit = len(keys)
	}
	var keyColors GradientTable
	fmt.Println(reverseColorCounter)
	fmt.Println(keys)
	for i, val := range keys {
		if i >= limit {
			break
		}
		cs := gwc.HexToRGB(gwc.CSS3NamesToHex[reverseColorCounter[val]])
		c := colorful.Color{float64(cs[0]) / 256, float64(cs[1]) / 256, float64(cs[2]) / 256}
		fmt.Println(c)
		keyColors = append(keyColors, struct {
			Col colorful.Color
			Pos float64
		}{Col: c, Pos: float64(i) / float64(limit)})
	}
	return keyColors
}
