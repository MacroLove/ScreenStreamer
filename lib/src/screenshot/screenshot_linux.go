package screenshot

import (
	"fmt"
	"image"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type POS struct {
	X, Y int
}

type SIZE struct {
	W, H int
}

type RESIZE struct {
	W, H int
}

type CAPTURE struct {
	W, H int
	B    *[]byte
}

func ScreenRect() (image.Rectangle, error) {
	c, err := xgb.NewConn()
	if err != nil {
		return image.Rectangle{}, err
	}
	defer c.Close()

	screen := xproto.Setup(c).DefaultScreen(c)
	x := screen.WidthInPixels
	y := screen.HeightInPixels

	return image.Rect(0, 0, int(x), int(y)), nil
}

func CaptureScreen() (*image.RGBA, error) {
	r, e := ScreenRect()
	if e != nil {
		return nil, e
	}
	return CaptureRect(r)
}

func CaptureRect(rect image.Rectangle) (*image.RGBA, error) {
	c, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}
	defer c.Close()

	screen := xproto.Setup(c).DefaultScreen(c)
	x, y := rect.Dx(), rect.Dy()
	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root), int16(rect.Min.X), int16(rect.Min.Y), uint16(x), uint16(y), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
	}

	img := &image.RGBA{data, 4 * x, image.Rect(0, 0, x, y)}
	return img, nil
}

func CaptureWindow(pos *POS, size *SIZE, resize *RESIZE) (*image.RGBA, error) {
	c, err := xgb.NewConn()
	if err != nil {
		fmt.Errorf("error occurred, when xgb.NewConn err:%v.\n", err)
	}
	defer c.Close()

	screen := xproto.Setup(c).DefaultScreen(c)

	aname := "_NET_ACTIVE_WINDOW"
	activeAtom, err := xproto.InternAtom(c, true, uint16(len(aname)), aname).Reply()
	if err != nil {
		fmt.Errorf("error occurred, when xproto.InternAtom 0 err:%v.\n", err)
	}

	reply, err := xproto.GetProperty(c, false, screen.Root, activeAtom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		fmt.Errorf("error occurred, when xproto.GetProperty 0 err:%v.\n", err)
	}
	windowId := xproto.Window(xgb.Get32(reply.Value))

	ginfo, err := xproto.GetGeometry(c, xproto.Drawable(windowId)).Reply()
	if err != nil {
		return nil, err
	}

	width := int(ginfo.Width) - pos.X
	height := int(ginfo.Height) - pos.Y
	if size.W != 0 && size.H != 0 {
		width = size.W
		height = size.H
	}

	xImg, err := xproto.GetImage(c, xproto.ImageFormatZPixmap, xproto.Drawable(windowId), int16(pos.X), int16(pos.Y), uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	data := xImg.Data
	for i := 0; i < len(data); i += 4 {
		data[i], data[i+2], data[i+3] = data[i+2], data[i], 255
	}

	img := &image.RGBA{data, 4 * width, image.Rect(pos.X, pos.Y, width, height)}
	return img, nil
}

func CaptureWindowMust(pos *POS, size *SIZE, resize *RESIZE) *image.RGBA {
	img, err := CaptureWindow(pos, size, resize)
	for err != nil {
		img, err = CaptureWindow(pos, size, resize)
	}
	return img
}
