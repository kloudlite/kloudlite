package main

import (
	"github.com/signintech/gopdf"
)

const pagePadding = 40.0

var effectiveWidth = gopdf.PageSizeA4.W - pagePadding*2

func main() {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{
		PageSize: *gopdf.PageSizeA4,
	})
	pdf.AddPage()
	pdf.AddTTFFont("OpenSans-Regular", "./fonts/OpenSans-Regular.ttf")
	pdf.AddTTFFont("OpenSans-Medium", "./fonts/OpenSans-Medium.ttf")
	pdf.AddTTFFont("OpenSans-Bold", "./fonts/OpenSans-Bold.ttf")

	base := func() {
		pdf.SetFont("OpenSans-Regular", "", 12)
		pdf.SetTextColor(71, 85, 105)
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

	pdf.AddOutline("Kloudlite Bill")
	pdf.WritePdf("hello.pdf")
}

func fromAdress(pdf *gopdf.GoPdf) {
	pdf.SetFont("OpenSans-Bold", "", 8)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetY(pdf.GetY() + 20)
	pdf.MultiCell(&gopdf.Rect{
		W: gopdf.PageSizeA4.W,
		H: 20,
	}, "From")
	pdf.SetFont("OpenSans-Regular", "", 10)
	wrap, _ := pdf.SplitTextWithWordWrap("Kloudlite Inc, \n651 N Broad St, Suite 201, Middletown, DE 19709 US", gopdf.PageSizeA4.W/3)
	for _, line := range wrap {
		pdf.MultiCell(&gopdf.Rect{
			W: gopdf.PageSizeA4.W / 3,
			H: 80,
		}, line)
	}
}

func summary(pdf *gopdf.GoPdf) {
	pdf.SetX(pagePadding)
	pdf.SetY(pagePadding + 200)
	pdf.SetFont("OpenSans-Medium", "", 12)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth,
		H: 20,
	}, "Summary")
	pdf.SetStrokeColor(203, 213, 225)
	pdf.SetLineType("dotted")
	pdf.SetLineWidth(0.5)
	pdf.Line(pagePadding, pdf.GetY()+3, gopdf.PageSizeA4.W-pagePadding, pdf.GetY()+3)
	pdf.SetY(pdf.GetY() + 10)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Total usage charges")
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "$444", gopdf.CellOption{
		Align: gopdf.Right,
	})
	pdf.Line(pagePadding, pdf.GetY()+25, gopdf.PageSizeA4.W-pagePadding, pdf.GetY()+25)
	pdf.SetFont("OpenSans-Bold", "", 14)
	pdf.SetX(pagePadding)
	pdf.SetY(pdf.GetY() + 35)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Total due")
	pdf.CellWithOption(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "$444", gopdf.CellOption{
		Align: gopdf.Right,
	})
	pdf.Br(10)
	pdf.SetFont("OpenSans-Regular", "", 8)
	pdf.SetX(pagePadding)
	pdf.SetY(pdf.GetY() + 10)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "If you have a credit card on file, it will be automatically charged within 24 hours")

	pdf.SetFont("OpenSans-Regular", "", 12)
	pdf.Cell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 30,
	}, "Product usage charges")

}

func details(pdf *gopdf.GoPdf) {
	pdf.SetFont("OpenSans-Bold", "", 8)
	pdf.SetTextColor(71, 85, 105)
	pdf.SetX(effectiveWidth/2 + pagePadding)
	pdf.SetY(pagePadding + 60)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth / 2,
		H: 20,
	}, "Details")
	pdf.SetFont("OpenSans-Regular", "", 10)

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
	pdf.SetTextColor(71, 85, 105)
	pdf.SetY(pdf.GetY() + 20)
	pdf.MultiCell(&gopdf.Rect{
		W: effectiveWidth/2 + pagePadding,
		H: 20,
	}, "For")
	pdf.SetFont("OpenSans-Regular", "", 10)
	wrap, _ := pdf.SplitTextWithWordWrap("Plaxonic Technologies Pvt Ltd, \n141 Stevens Ave STE 5 Oldsmar\nFL 34677 US", gopdf.PageSizeA4.W/3)
	for _, line := range wrap {
		pdf.MultiCell(&gopdf.Rect{
			W: gopdf.PageSizeA4.W / 3,
			H: 80,
		}, line)
	}
}

func header(pdf *gopdf.GoPdf) {
	pdf.Image("./images/logo-scalable.png", pagePadding, pagePadding, &gopdf.Rect{
		W: 323 / 3,
		H: 51 / 3,
	})
	pdf.SetY(pagePadding + 25)
	pdf.SetX(pagePadding)
	pdf.SetFont("OpenSans-Medium", "", 12)
	pdf.MultiCell(&gopdf.Rect{
		W: gopdf.PageSizeA4.W,
		H: 20,
	}, "Final invoice for the January 2022 billing period")
}
