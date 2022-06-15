package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
)

var (
	out     = flag.String("o", "output", "output file path")
	outtype = flag.String("t", "png", "output file type. png webp jpg")
	prefix  = flag.String("p", "tp-", "css class name prefix")
	t       = template.Must(template.New("css").Parse(`/* ----------------------------------------------------
   created with https://github.com/nzlov/tp
   ----------------------------------------------------

   usage: <span class="{-spritename-} sprite"></span>

   replace {-spritename-} with the sprite you like to use

   ----------------------------------------------------
*/

.sprite {display:inline-block; overflow:hidden; background-repeat: no-repeat;background-image:url(./{{.Name}}.{{.OutType}});}
{{$prefix := .Prefix}}
{{range .Sprites}}
.{{$prefix}}{{.Name}} {width:{{.W}}px; height:{{.H}}px; background-position: -{{.X}}px -{{.Y}}px}
{{end}}
  `))
)

func main() {
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("need src path")
		return
	}
	is, err := loadimage(flag.Arg(0))
	if err != nil {
		panic(err)
	}
	if err := save(*out, is); err != nil {
		panic(err)
	}
}

func save(name string, is []*Item) error {
	img, err := flow(is)
	if err != nil {
		return err
	}

	cf, err := os.Create(name + ".css")
	if err != nil {
		return err
	}
	defer cf.Close()
	if err := t.Execute(cf, map[string]any{
		"Prefix":  prefix,
		"OutType": outtype,
		"Name":    name,
		"Sprites": is,
	}); err != nil {
		return err
	}
	f, err := os.Create(name + "." + *outtype)
	if err != nil {
		return err
	}
	defer f.Close()
	switch *outtype {
	case "jpg":
		return jpeg.Encode(f, img, &jpeg.Options{Quality: 100})
	case "webp":
		options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 100)
		if err != nil {
			return err
		}
		return webp.Encode(f, img, options)
	default:
		return png.Encode(f, img)
	}
}

func flow(is []*Item) (image.Image, error) {
	// 计算长宽
	mw := 0
	mh := 0
	for _, v := range is {
		if v.H > mh {
			mh = v.H
		}
		mw += v.W
	}
	img := image.NewRGBA(image.Rect(0, 0, mw, mh))
	// 横放
	cx := 0
	for _, v := range is {
		v.X = cx
		cx = v.X + v.W
		draw.Draw(img, image.Rect(v.X, v.Y, cx, v.H), v.img, image.Point{0, 0}, draw.Over)
	}
	return img, nil
}

func loadimage(path string) ([]*Item, error) {
	is := []*Item{}
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".png") || strings.HasSuffix(path, ".jpg") {
			fmt.Println("Load:", path)
			i, err := NewItem(path)
			if err != nil {
				return err
			}
			is = append(is, i)
		}
		return nil
	})

	return is, err
}

type Item struct {
	Name   string
	format string
	W, H   int
	X, Y   int
	img    image.Image
}

func NewItem(path string) (*Item, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	r := img.Bounds()
	return &Item{
		Name:   filename(f.Name()),
		format: format,
		img:    img,
		W:      r.Dx(),
		H:      r.Dy(),
	}, nil
}

func filename(path string) string {
	s := strings.LastIndex(path, "/") + 1
	e := strings.LastIndex(path, ".")
	return path[s:e]
}
