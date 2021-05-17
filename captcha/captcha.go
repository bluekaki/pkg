package captcha

import (
	"bytes"
	crand "crypto/rand"
	"encoding/pem"
	"io"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/bluekaki/pkg/errors"

	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)

const (
	wordLen       = 6
	channelBuffer = 10
	alphabet      = "0123456789abcdefghijklmnopqrstuvwxyz"
)

var defaultFont *truetype.Font

func init() {
	block, _ := pem.Decode([]byte(deliusUnicase))

	var err error
	defaultFont, err = truetype.Parse(block.Bytes)
	if err != nil {
		panic(err)
	}
}

// Captcha a 6 words png captcha
type Captcha interface {
	Simple() (code string, raw []byte)
}

type img struct {
	code string
	raw  []byte
}

type captcha struct {
	width     float64
	height    float64
	alphabet  [len(alphabet)]string
	simplePNG chan *img
}

// New create a generator
func New(width, height uint8) (Captcha, error) {
	if width == 0 {
		return nil, errors.New("width must be bigger than zero")
	}
	if height == 0 {
		return nil, errors.New("height must be bigger than zero")
	}

	captcha := &captcha{
		width:     float64(width),
		height:    float64(height),
		simplePNG: make(chan *img, channelBuffer),
	}

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i, index := range random.Perm(len(alphabet)) {
		captcha.alphabet[i] = string(alphabet[index])
	}

	go captcha.simple()
	return captcha, nil
}

func (c *captcha) simple() {
	var random *rand.Rand
	resetRandom := func() {
		random = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	counter := uint64(0)
	for {
		if counter++; counter%20 == 1 {
			resetRandom()
		}

		dc := gg.NewContext(int(c.width), int(c.height))
		dc.SetRGB(1, 1, 1)
		dc.Clear()

		{ // draw lines
			for i := 0; i < 10; i++ {
				x1 := random.Float64() * c.width
				y1 := random.Float64() * c.height
				x2 := random.Float64() * c.width
				y2 := random.Float64() * c.height
				r := random.Float64()
				g := random.Float64()
				b := random.Float64()
				a := random.Float64()*0.5 + 0.5
				w := random.Float64()*4 + 1
				dc.SetRGBA(r, g, b, a)
				dc.SetLineWidth(w)
				dc.DrawLine(x1, y1, x2, y2)
				dc.Stroke()
			}
		}

		dc.SetRGB(0, 0, 0)

		seed := make([]byte, wordLen)
		io.ReadFull(crand.Reader, seed)

		words := make([]string, wordLen)
		for i, v := range seed {
			words[i] = c.alphabet[int(v)%len(alphabet)]
		}

		for i, word := range words {
			face := truetype.NewFace(defaultFont, &truetype.Options{
				Size: float64(random.Intn(int(c.height*0.6))) + c.height*0.4,
			})

			dc.SetFontFace(face)
			dc.Stroke()

			degree := float64(random.Intn(45) + 1)
			β := gg.Radians(degree)

			if i%2 == 0 {
				degree = -degree
			}

			x := float64(i*int(c.width)/wordLen) + c.width*0.1
			y := float64(face.Metrics().Height.Round() / 2)

			var x1, y1 float64
			if degree < 0 {
				x1 = x*math.Cos(β) - y*math.Sin(β)
				y1 = y*math.Cos(β) + x*math.Sin(β)

			} else {
				x1 = x*math.Cos(β) + y*math.Sin(β)
				y1 = y*math.Cos(β) - x*math.Sin(β)
			}

			dc.Rotate(gg.Radians(degree))
			dc.DrawStringAnchored(word, x1, y1, 0.5, 0.5)
			dc.Rotate(gg.Radians(-degree))
		}

		buf := bytes.NewBuffer(nil)
		dc.EncodePNG(buf)

		c.simplePNG <- &img{code: strings.Join(words, ""), raw: buf.Bytes()}
	}
}

func (c *captcha) Simple() (code string, raw []byte) {
	img := <-c.simplePNG
	return img.code, img.raw
}
