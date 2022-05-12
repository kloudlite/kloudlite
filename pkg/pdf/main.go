package main

import (
	"fmt"

	"github.com/signintech/gopdf"
)

const topMargin = 40
const leftMargin = 35.0

var effectiveWidth = gopdf.PageSizeA4.W - leftMargin*2

type Table struct {
	Header []string
	rows   [][]string
}

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.AddPage()
	pdf.AddTTFFont("OpenSans-Regular", "./fonts/OpenSans-Regular.ttf")
	pdf.AddTTFFont("OpenSans-Medium", "./fonts/OpenSans-Medium.ttf")
	pdf.AddTTFFont("OpenSans-Bold", "./fonts/OpenSans-Bold.ttf")
	pdf.AddTTFFont("OpenSans-Italic", "./fonts/OpenSans-Italic.ttf")

	base := func() {
		pdf.SetFont("OpenSans-Regular", "", 12)
		pdf.SetTextColor(71, 85, 105)
		// pdf.SetTextColor(255, 0, 0)
	}

	base()
	header(&pdf)

	base()
	fromAdress(&pdf)

	base()
	forAddress(&pdf)

	base()
	details(&pdf)

	base()
	summary(&pdf)

	// pdf.SetY(pdf.GetY() + 55)
	tableSection(&pdf, &Table{
		Header: []string{"Droplets", "Hours", "Start", "End", "$29.81"},
		rows: [][]string{
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},

			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
		},
	})

	tableSection(&pdf, &Table{
		Header: []string{"Droplets", "Hours", "Start", "End", "$29.81"},
		rows: [][]string{
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},

			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"remote-baby(s-4vcpu-8gb-amd)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
			{"plaxlb (s-vcpu-1gb)", "744", "01-01 00:00", "01-01 00:00", "$29.81"},
		},
	})

	setPage(&pdf)

	pdf.AddOutline("Kloudlite Bill")
	pdf.WritePdf("hello.pdf")
}

func setPage(goPdf *gopdf.GoPdf) {
	goPdf.SetFont("OpenSans-Medium", "", 9)

	goPdf.SetX(gopdf.PageSizeA4.W - leftMargin - effectiveWidth/2)
	goPdf.SetY(gopdf.PageSizeA4.H - topMargin)

	goPdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, fmt.Sprintf("Page %v", goPdf.GetNumberOfPages()), gopdf.CellOption{
		Align: gopdf.Right,
	})
}

func tableSection(goPdf *gopdf.GoPdf, table *Table) {
	tableTop := goPdf.GetY() + 30

	if tableTop >= gopdf.PageSizeA4.H-topMargin*2 {
		setPage(goPdf)
		goPdf.AddPage()
		tableTop = topMargin
	}

	titleText, hoursText, startText, endText, priceText := table.Header[0], table.Header[1], table.Header[2], table.Header[3], table.Header[4]

	goPdf.SetFont("OpenSans-Bold", "", 8)
	// goPdf.SetTextColor(71, 85, 105)
	goPdf.SetY(tableTop + 2)
	goPdf.SetX(leftMargin)
	goPdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + leftMargin,
		H: 20,
	}, titleText)

	goPdf.SetY(tableTop + 2)
	goPdf.SetX(leftMargin + 277)

	goPdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + leftMargin,
		H: 20,
	}, hoursText)

	goPdf.SetY(tableTop + 2)
	goPdf.SetX(leftMargin + 326)

	goPdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + leftMargin,
		H: 20,
	}, startText)

	goPdf.SetY(tableTop + 2)
	goPdf.SetX(leftMargin + 394)

	goPdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + leftMargin,
		H: 20,
	}, endText)

	goPdf.SetY(tableTop + 2)
	goPdf.SetX(gopdf.PageSizeA4.W - effectiveWidth/2 - leftMargin)

	goPdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, priceText, gopdf.CellOption{
		Align: gopdf.Right,
	})

	goPdf.SetY(goPdf.GetY() + 19)

	goPdf.Line(leftMargin, goPdf.GetY(), gopdf.PageSizeA4.W-leftMargin, goPdf.GetY())

	goPdf.SetY(goPdf.GetY() + 6)

	// for in table.rows
	for _, row := range table.rows {

		titleText, hoursText, startText, endText, priceText := row[0], row[1], row[2], row[3], row[4]

		tableTop = goPdf.GetY()

		if tableTop >= gopdf.PageSizeA4.H-topMargin-15 {
			setPage(goPdf)
			goPdf.AddPage()
			tableTop = topMargin
		}

		goPdf.SetFont("OpenSans-Regular", "", 8)
		// goPdf.SetTextColor(71, 85, 105)
		goPdf.SetY(tableTop + 2)
		goPdf.SetX(leftMargin)
		goPdf.MultiCell(&gopdf.Rect{
			W: effectiveWidth/2 + leftMargin,
			H: 20,
		}, titleText)

		goPdf.SetY(tableTop + 2)
		goPdf.SetX(leftMargin + 277)

		goPdf.MultiCell(&gopdf.Rect{
			W: effectiveWidth/2 + leftMargin,
			H: 20,
		}, hoursText)

		goPdf.SetY(tableTop + 2)
		goPdf.SetX(leftMargin + 326)

		goPdf.MultiCell(&gopdf.Rect{
			W: effectiveWidth/2 + leftMargin,
			H: 20,
		}, startText)

		goPdf.SetY(tableTop + 2)
		goPdf.SetX(leftMargin + 394)

		goPdf.MultiCell(&gopdf.Rect{
			W: effectiveWidth/2 + leftMargin,
			H: 20,
		}, endText)

		goPdf.SetY(tableTop + 2)
		goPdf.SetX(gopdf.PageSizeA4.W - effectiveWidth/2 - leftMargin)

		goPdf.CellWithOption(&gopdf.Rect{
			W: effectiveWidth / 2,
			H: 30,
		}, priceText, gopdf.CellOption{
			Align: gopdf.Right,
		})

		goPdf.SetY(goPdf.GetY() + 19)
	}
}

func fromAdress(pdf *gopdf.GoPdf) {
	pdf.SetFont("OpenSans-Bold", "", 8)
	// pdf.SetTextColor(71, 85, 105)
	pdf.SetY(pdf.GetY() + 31.5)
	pdf.MultiCell(&gopdf.Rect{
		W: gopdf.PageSizeA4.W,
		H: 20,
	}, "From")

	pdf.SetY(pdf.GetY() + 3.9)
	pdf.SetFont("OpenSans-Regular", "", 8.5)

	wrap, _ := pdf.SplitTextWithWordWrap("Kloudlite Inc, \n651 N Broad St, Suite 201, Middletown, \nDE 19709 US", gopdf.PageSizeA4.W/3)
	for _, line := range wrap {
		pdf.MultiCell(&gopdf.Rect{
			W: gopdf.PageSizeA4.W / 3,
			H: 80,
		}, line)

		pdf.SetY(pdf.GetY() + -1)
	}
}

func summary(pdf *gopdf.GoPdf) {
	pdf.SetX(leftMargin)
	pdf.SetY(topMargin + 245)

	pdf.SetFont("OpenSans-Medium", "", 12)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth,
		H: 20,
	}, "Summary")

	pdf.SetY(pdf.GetY() + 8)
	pdf.SetStrokeColor(203, 213, 225)
	// pdf.SetLineType("dotted")
	pdf.SetLineWidth(0.5)
	pdf.Line(leftMargin, pdf.GetY(), gopdf.PageSizeA4.W-leftMargin, pdf.GetY())

	pdf.SetY(pdf.GetY() + 8)

	pdf.SetFont("OpenSans-Regular", "", 8.5)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Total usage charges")

	pdf.SetFont("OpenSans-Medium", "", 8.5)
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "$444.89", gopdf.CellOption{
		Align: gopdf.Right,
	})

	pdf.SetY(pdf.GetY() + 19)

	pdf.SetStrokeColor(203, 213, 225)
	pdf.Line(leftMargin, pdf.GetY(), gopdf.PageSizeA4.W-leftMargin, pdf.GetY())

	pdf.SetFont("OpenSans-Bold", "", 12.5)
	pdf.SetX(leftMargin)
	pdf.SetY(pdf.GetY() + 9)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Total due")

	pdf.SetFont("OpenSans-Bold", "", 12.5)
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "$444.89", gopdf.CellOption{
		Align: gopdf.Right,
	})

	pdf.SetFont("OpenSans-Regular", "", 8)
	pdf.SetX(leftMargin)

	pdf.SetY(pdf.GetY() + 25)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "If you have a credit card on file, it will be automatically charged within 24 hours")

	pdf.SetY(pdf.GetY() + 39)
	pdf.Line(leftMargin, pdf.GetY(), gopdf.PageSizeA4.W-leftMargin, pdf.GetY())

	pdf.SetY(pdf.GetY() + 23)
	pdf.SetX(leftMargin)
	pdf.SetFont("OpenSans-Regular", "", 12.5)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Product usage charges")

	pdf.SetY(pdf.GetY() + 19.5)
	pdf.SetX(leftMargin)
	pdf.SetFont("OpenSans-Italic", "", 8.75)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Detailed usage infromation is avaailable via the API or can be downloaded from the billing section of your account")

	pdf.SetY(pdf.GetY() + 15)

}

func details(pdf *gopdf.GoPdf) {
	pdf.SetFont("OpenSans-Bold", "", 8)
	// pdf.SetTextColor(71, 85, 105)
	pdf.SetX(effectiveWidth/2 + leftMargin)
	pdf.SetY(topMargin + 81)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 20,
	}, "Details")

	pdf.SetY(pdf.GetY() + 6.5)

	pdf.SetFont("OpenSans-Regular", "", 8.5)

	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "Invoice Number")

	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "123456789", gopdf.CellOption{
		Align: gopdf.Right,
	})
	pdf.Br(14)
	pdf.SetX(gopdf.PageSizeA4.W / 2)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "Date of issue:")
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "February 1, 2022", gopdf.CellOption{
		Align: gopdf.Right,
	})
	pdf.Br(14)
	pdf.SetX(gopdf.PageSizeA4.W / 2)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "Payment due on:")
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 4,
		H: 100,
	}, "February 1, 2022", gopdf.CellOption{
		Align: gopdf.Right,
	})
}

func forAddress(pdf *gopdf.GoPdf) {
	pdf.SetFont("OpenSans-Bold", "", 8)
	// pdf.SetTextColor(71, 85, 105)
	pdf.SetY(pdf.GetY() + 15.5)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + leftMargin,
		H: 20,
	}, "For")

	pdf.SetY(pdf.GetY() + 5)
	pdf.SetFont("OpenSans-Regular", "", 8)
	wrap, _ := pdf.SplitTextWithWordWrap("Plaxonic Technologies Pvt Ltd, \n141 Stevens Ave STE 5 Oldsmar\nFL 34677 US", gopdf.PageSizeA4.W/3)
	for _, line := range wrap {
		pdf.MultiCell(&gopdf.Rect{
			W: gopdf.PageSizeA4.W / 3,
			H: 80,
		}, line)
	}
}

func header(pdf *gopdf.GoPdf) {
	// pdf.Image("./images/template.png", 0, 0, &gopdf.Rect{
	// 	W: gopdf.PageSizeA4.W,
	// 	H: gopdf.PageSizeA4.H,
	// })

	const imageRatioDevider = 2.2
	pdf.Image("./images/logo-scalable.png", leftMargin, topMargin, &gopdf.Rect{
		W: 323 / imageRatioDevider,
		H: 51 / imageRatioDevider,
	})

	pdf.SetY(topMargin + 32)
	pdf.SetX(leftMargin)
	pdf.SetFont("OpenSans-Medium", "", 12.4)
	pdf.MultiCell(&gopdf.Rect{
		W: gopdf.PageSizeA4.W,
		H: 20,
	}, "Final invoice for the January 2022 billing period")
}
