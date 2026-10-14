# Barcode.wasm for Go

Go から使えるバーコード生成ライブラリです。バーコードを作る本体は C++ の WebAssembly（`wasm/barcode.wasm`）で、Go のコードは Node.js を通してそれを呼び出します。

- 対応：CODE39・CODE93・CODE128・GS1-128・NW7・ITF・Matrix 2of5・NEC 2of5・JAN-8・JAN-13・UPC-A・UPC-E・GS1 DataBar（標準・限定・拡張）・郵便カスタマーバーコード・QR コード・DataMatrix・PDF417
- 出力：PNG・JPEG・SVG（data URI の文字列）
- **体験版**です。画像に赤い「SAMPLE」が入ります。製品版は [Barcode.wasm](https://www.pao.ac/barcode.wasm/) をご覧ください。

## 必要なもの

- Go 1.21 以降
- Node.js 18 以降（`node` が PATH にあること）

## インストール

```bash
go get github.com/barcode-wasm/go@v1.2.2
```

## 使い方

```go
package main

import (
	"fmt"

	barcode "github.com/barcode-wasm/go"
)

func main() {
	c := barcode.NewCode128(barcode.FormatPNG)
	c.SetShowText(true)
	img, err := c.Draw("ABC123", 400, 120) // "data:image/png;base64,..."
	if err != nil {
		panic(err)
	}
	fmt.Println(img[:40])

	qr := barcode.NewQRCode(barcode.FormatSVG)
	svg, _ := qr.Draw("https://www.pao.ac/", 300)
	fmt.Println(len(svg))
}
```

PNG に保存する例は [examples/code128](examples/code128/main.go) にあります（`go run ./examples/code128`）。

## WASM ファイルの場所

`wasm/` フォルダーは、モジュールのソースの隣（`go get` したモジュールキャッシュ）から自動で見つけます。`-trimpath` でビルドした場合や、実行ファイルを別の場所へ配る場合は、`wasm/` フォルダーを実行ファイルと同じ場所に置いてください。

## ライセンス

MIT（[LICENSE](LICENSE)）

---

# Barcode.wasm for Go (English)

A barcode generation library for Go. The barcode engine is a C++ WebAssembly module (`wasm/barcode.wasm`); the Go code calls it through Node.js.

- Symbologies: CODE39, CODE93, CODE128, GS1-128, NW7 (Codabar), ITF, Matrix 2of5, NEC 2of5, JAN-8/EAN-8, JAN-13/EAN-13, UPC-A, UPC-E, GS1 DataBar (Omnidirectional, Limited, Expanded), Japan Post customer barcode, QR Code, DataMatrix, PDF417
- Output: PNG, JPEG, SVG (as a data URI string)
- This is a **trial edition**: images carry a red "SAMPLE" mark. See [Barcode.wasm](https://www.pao.ac/barcode.wasm/) for the product edition.

## Requirements

- Go 1.21 or later
- Node.js 18 or later (`node` on PATH)

## Install

```bash
go get github.com/barcode-wasm/go@v1.2.2
```

## Usage

```go
c := barcode.NewCode128(barcode.FormatPNG)
c.SetShowText(true)
img, err := c.Draw("ABC123", 400, 120) // "data:image/png;base64,..."
```

See [examples/code128](examples/code128/main.go) for saving PNG files (`go run ./examples/code128`).

## Where the WASM files are found

The `wasm/` folder is located next to the module source (in the Go module cache). If you build with `-trimpath` or ship the executable elsewhere, place the `wasm/` folder next to the executable.

## License

MIT ([LICENSE](LICENSE))
