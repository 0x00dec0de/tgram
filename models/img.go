package models

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"regexp"
	"strings"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	_ "golang.org/x/image/webp"

	_ "golang.org/x/image/bmp"

	"github.com/MaxHalford/halfgone"
	"github.com/nfnt/resize"
	sp "github.com/recoilme/slowpoke"
	"github.com/recoilme/tgram/utils"
)

const (
	// lang/username/img
	dbImg     = "img/%s/%s/%d.png"
	dbImgOrig = "img/%s/%s/o%d.png"

	// lang/username/counter
	dbImgID = "img/%s/%s/id"

	// max size
	maxSize = 10240000
)

// ImgProcess extract img links from markdown, download, and replace with local copy
func ImgProcess(s, lang, username string) (res string, err error) {
	//fmt.Printf("s:'%s'", s)
	r, err := regexp.Compile(`!\[(.*?)\]\((.*?)\)`)
	if err != nil {
		return s, err
	}
	//res = r.ReplaceAllStringFunc(s, isImg)
	//r.FindAllString(s,)
	submatchall := r.FindAllString(s, -1)
	for _, element := range submatchall {
		b, href := isImg(element)
		if b != nil {
			store(href, lang, username, b)
		}
	}
	return res, err
}

func store(href, lang, username string, b []byte) {
	//image processing
	if img, _, err := image.Decode(bytes.NewReader(b)); err == nil {
		small := resize.Thumbnail(800, 800, img, resize.MitchellNetravali)
		smallb := new(bytes.Buffer)
		if err := png.Encode(smallb, small); err == nil {
			ad := halfgone.AtkinsonDitherer{}.Apply(halfgone.ImageToGray(small))
			//store
			if imgid, err := sp.Counter(fmt.Sprintf(dbImgID, lang, username), []byte("id")); err == nil {
				f := fmt.Sprintf(dbImg, lang, username, imgid)
				defer sp.Close(f)
				grayb := new(bytes.Buffer)
				err = png.Encode(grayb, ad)
				if err := sp.Set(f, Uint32toBin(uint32(imgid)), grayb.Bytes()); err == nil {
					fo := fmt.Sprintf(dbImgOrig, lang, username, imgid)
					sp.Set(fo, Uint32toBin(uint32(imgid)), smallb.Bytes())
					defer sp.Close(fo)
				}

			}

		}
	}

}

func isImg(s string) ([]byte, string) {
	//fmt.Println("img:", s)
	var href = ""
	//var err error
	first := strings.IndexByte(s, '(') + 1
	last := strings.IndexByte(s, ')')
	if first > 0 && last > 0 && last > first {
		// extract link
		href = s[first:last]
		len := utils.HTTPImgLen(href)
		//log.Println("href", href, "len", len)
		if len > 0 && len < maxSize {
			//try download
			b := utils.HTTPGetBody(href)
			if b != nil {
				return b, href

			} else {
				//log.Println("b is nil")
			}
		}
		//fmt.Println(href)
		//href = "http://ya.ru"
		//s = s[:first] + href + s[last:]
		//fmt.Println(s)
	}

	return nil, ""
}

// convertToPNG converts from any recognized format to PNG.
func convertToPNG(w io.Writer, r io.Reader) error {
	img, _, err := image.Decode(r)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}
