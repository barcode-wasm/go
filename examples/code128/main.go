// Example: draw a CODE128 barcode and a QR code and save them as PNG files.
//
//	go run ./examples/code128
package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	barcode "github.com/barcode-wasm/go"
)

func save(name, dataURI string) {
	b64 := dataURI[strings.Index(dataURI, ",")+1:]
	png, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(name, png, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("saved", name)
}

func main() {
	c128 := barcode.NewCode128(barcode.FormatPNG)
	c128.SetShowText(true)
	img, err := c128.Draw("ABC123", 400, 120)
	if err != nil {
		log.Fatal(err)
	}
	save("code128.png", img)

	qr := barcode.NewQRCode(barcode.FormatPNG)
	img, err = qr.Draw("https://www.pao.ac/", 300)
	if err != nil {
		log.Fatal(err)
	}
	save("qr.png", img)
}
