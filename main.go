package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"sort"

	"gocv.io/x/gocv"
)

const imgPath = "/home/rudy/Images/IMG_20210210_122855.jpg"

func main() {
	img := gocv.IMRead(imgPath, gocv.IMReadAnyColor)
	if img.Empty() {
		log.Fatal("can't read image file")
	}
	// rescale original image
	// gocv.Resize(img, &img, image.Point{}, 0.3, 0.3, gocv.InterpolationDefault)
	defer img.Close()

	oImg := img.Clone()
	defer oImg.Close()

	// convert image to greyscale
	gImg := gocv.NewMat()
	gocv.CvtColor(img, &gImg, gocv.ColorBGRToGray)
	// adjut contrast
	// gImg.ConvertToWithParams(&gImg, gImg.Type(), 0.6, 0)

	// reduce image noise
	nImg := gocv.NewMat()
	gocv.BilateralFilter(gImg, &nImg, 11, 17, 17)
	win := gocv.NewWindow("greyscale without noise")
	win.IMShow(nImg)
	gImg.Close()

	// run edge detection algorithm
	eImg := gocv.NewMat()
	gocv.Canny(nImg, &eImg, 30, 200)
	win = gocv.NewWindow("edge detection")
	win.IMShow(eImg)
	nImg.Close()

	// calculate contours
	contours := gocv.FindContours(eImg, gocv.RetrievalList, gocv.ChainApproxSimple)
	contours = sortContoursByArea(contours)
	gocv.DrawContours(&eImg, contours, -1, color.RGBA{0, 0, 255, 255}, 3)
	win = gocv.NewWindow("Test opencv")
	win.IMShow(eImg)

	// find the biggest contour with 4 sides
	plate := gocv.PointVector{}
	for i := 0; i < contours.Size(); i++ {
		cc := contours.At(contours.Size() - i - 1)
		p := gocv.ArcLength(cc, true)
		approx := gocv.ApproxPolyDP(cc, 0.018*p, true)
		if approx.Size() == 4 {
			fmt.Println("Found plate!!!")
			plate = cc
			break
		}
	}
	if plate.IsNil() {
		log.Fatal("Unable to find any plate in this picture")
	}

	nContours := gocv.NewPointsVector()
	nContours.Append(plate)
	gocv.DrawContours(&oImg, nContours, -1, color.RGBA{0, 0, 255, 255}, 3)
	rect := gocv.BoundingRect(plate)
	rect.Max.Y -= 20
	rect.Max.X -= 10
	rect.Min.X += 15
	rect.Min.Y += 20
	r := oImg.Region(rect)
	gocv.GaussianBlur(r, &r, image.Point{175, 175}, 0, 0, gocv.BorderDefault)
	r.Close()

	win = gocv.NewWindow("Test opencv 2")
	win.IMShow(oImg)
	win.WaitKey(0)
}

func sortContoursByArea(c gocv.PointsVector) gocv.PointsVector {
	n := gocv.NewPointsVector()
	var indexes []int
	for i := 0; i < c.Size(); i++ {
		indexes = append(indexes, i)
	}
	sort.Slice(indexes, func(i, j int) bool {
		a1 := gocv.ContourArea(c.At(indexes[i]))
		a2 := gocv.ContourArea(c.At(indexes[j]))
		return a1 < a2
	})
	// keep only the last 30 elements
	for _, i := range indexes[int(math.Max(float64(c.Size()-30), 0)):] {
		n.Append(c.At(i))
	}
	return n
}
