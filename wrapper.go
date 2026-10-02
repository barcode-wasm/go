// Package barcode_pao_wasm provides Go wrappers for the barcode C++ WASM module.
// It uses Node.js as a bridge to execute the Emscripten-compiled WASM engine.
//
// Architecture:
//
//	Go code → JSON → node _barcode_runner.mjs → barcode.mjs → WASM → JSON → Go
package barcode_pao_wasm

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// Output format constants.
const (
	FormatPNG  = "png"
	FormatJPEG = "jpg"
	FormatSVG  = "svg"
)

// ─── Node.js detection ──────────────────────────────────────────────────────

var (
	nodePath     string
	nodePathOnce sync.Once
	nodePathErr  error
)

func findNode() (string, error) {
	nodePathOnce.Do(func() {
		for _, cmd := range []string{"node", "nodejs"} {
			p, err := exec.LookPath(cmd)
			if err == nil {
				nodePath = p
				return
			}
		}
		nodePathErr = fmt.Errorf("Node.js is required for WASM barcode generation. Please install Node.js from https://nodejs.org/")
	})
	return nodePath, nodePathErr
}

// ─── WASM directory resolution ──────────────────────────────────────────────

func getWasmDir() string {
	// 1. Try relative to this source file (development time)
	_, thisFile, _, ok := runtime.Caller(0)
	if ok {
		dir := filepath.Join(filepath.Dir(thisFile), "wasm")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	// 2. Try relative to executable (distribution)
	exePath, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exePath), "barcode_pao_wasm", "wasm")
		if info, err2 := os.Stat(dir); err2 == nil && info.IsDir() {
			return dir
		}
		dir = filepath.Join(filepath.Dir(exePath), "wasm")
		if info, err2 := os.Stat(dir); err2 == nil && info.IsDir() {
			return dir
		}
	}
	// 3. Fallback
	return "wasm"
}

// ─── WASM call protocol ─────────────────────────────────────────────────────

type wasmRequest struct {
	ClassName  string                 `json:"className"`
	Method     string                 `json:"method"`
	MethodArgs []interface{}          `json:"methodArgs"`
	Settings   map[string]interface{} `json:"settings"`
}

type wasmResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func callWasm(className, method string, args []interface{}, settings map[string]interface{}) (string, error) {
	node, err := findNode()
	if err != nil {
		return "", err
	}

	wasmDir := getWasmDir()
	scriptPath := filepath.Join(wasmDir, "_barcode_runner.mjs")

	req := wasmRequest{
		ClassName:  className,
		Method:     method,
		MethodArgs: args,
		Settings:   settings,
	}
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	cmd := exec.Command(node, scriptPath, string(reqJSON))
	cmd.Dir = wasmDir

	output, err := cmd.Output()
	if err != nil {
		// Try to extract error from stderr or stdout
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr := string(exitErr.Stderr)
			stdout := string(output)
			// Check stdout for JSON error
			if stdout != "" {
				var resp wasmResponse
				if jsonErr := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &resp); jsonErr == nil && resp.Error != "" {
					return "", fmt.Errorf("WASM error: %s", resp.Error)
				}
			}
			if stderr != "" {
				return "", fmt.Errorf("WASM execution failed: %s", strings.TrimSpace(stderr))
			}
		}
		return "", fmt.Errorf("WASM execution failed: %w", err)
	}

	// Find JSON line in output (may have debug output before it)
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var jsonLine string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			jsonLine = line
		}
	}

	if jsonLine == "" {
		return "", fmt.Errorf("no JSON response from WASM output")
	}

	var resp wasmResponse
	if err := json.Unmarshal([]byte(jsonLine), &resp); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}
	if resp.Error != "" {
		return "", fmt.Errorf("WASM error: %s", resp.Error)
	}

	result, ok := resp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type from WASM")
	}
	return result, nil
}

// ═════════════════════════════════════════════════════════════════════════════
// Base types
// ═════════════════════════════════════════════════════════════════════════════

// BarcodeWasmBase holds common settings for all barcode types.
type BarcodeWasmBase struct {
	className string
	settings  map[string]interface{}
}

func newBase(className string) BarcodeWasmBase {
	return BarcodeWasmBase{
		className: className,
		settings: map[string]interface{}{
			"outputFormat": FormatPNG,
		},
	}
}

// SetOutputFormat sets the output format (png, jpg, svg).
func (b *BarcodeWasmBase) SetOutputFormat(format string) {
	b.settings["outputFormat"] = format
}

// SetForegroundColor sets the foreground color (RGBA).
func (b *BarcodeWasmBase) SetForegroundColor(r, g, bl, a int) {
	b.settings["foregroundColor"] = []int{r, g, bl, a}
}

// SetBackgroundColor sets the background color (RGBA).
func (b *BarcodeWasmBase) SetBackgroundColor(r, g, bl, a int) {
	b.settings["backgroundColor"] = []int{r, g, bl, a}
}

func (b *BarcodeWasmBase) call(method string, args ...interface{}) (string, error) {
	return callWasm(b.className, method, args, b.settings)
}

// Barcode1DBase provides common 1D barcode settings.
type Barcode1DBase struct {
	BarcodeWasmBase
}

func new1DBase(className string) Barcode1DBase {
	return Barcode1DBase{BarcodeWasmBase: newBase(className)}
}

// SetShowText sets whether to show text below the barcode.
func (b *Barcode1DBase) SetShowText(show bool) { b.settings["showText"] = show }

// SetTextGap sets the gap between barcode and text.
func (b *Barcode1DBase) SetTextGap(gap float64) { b.settings["textGap"] = gap }

// SetTextFontScale sets the text font scale.
func (b *Barcode1DBase) SetTextFontScale(scale float64) { b.settings["textFontScale"] = scale }

// SetTextEvenSpacing sets text even spacing mode.
func (b *Barcode1DBase) SetTextEvenSpacing(even bool) { b.settings["textEvenSpacing"] = even }

// SetFitWidth sets whether to fit the barcode to width.
func (b *Barcode1DBase) SetFitWidth(fit bool) { b.settings["fitWidth"] = fit }

// SetPxAdjustBlack sets pixel adjustment for black bars.
func (b *Barcode1DBase) SetPxAdjustBlack(adj int) { b.settings["pxAdjustBlack"] = adj }

// SetPxAdjustWhite sets pixel adjustment for white bars.
func (b *Barcode1DBase) SetPxAdjustWhite(adj int) { b.settings["pxAdjustWhite"] = adj }

// Draw generates a 1D barcode and returns Base64 or SVG string.
func (b *Barcode1DBase) Draw(code string, width, height int) (string, error) {
	return b.call("draw", code, width, height)
}

// Barcode2DBase provides common 2D barcode settings.
type Barcode2DBase struct {
	BarcodeWasmBase
}

func new2DBase(className string) Barcode2DBase {
	return Barcode2DBase{BarcodeWasmBase: newBase(className)}
}

// SetStringEncoding sets the string encoding (utf-8, shift-jis).
func (b *Barcode2DBase) SetStringEncoding(enc string) { b.settings["stringEncoding"] = enc }

// SetFitWidth sets whether to fit the barcode to width.
func (b *Barcode2DBase) SetFitWidth(fit bool) { b.settings["fitWidth"] = fit }

// Draw generates a 2D barcode and returns Base64 or SVG string.
func (b *Barcode2DBase) Draw(code string, size int) (string, error) {
	return b.call("draw", code, size)
}

// ═════════════════════════════════════════════════════════════════════════════
// 1D Barcodes
// ═════════════════════════════════════════════════════════════════════════════

// Code39 generates Code39 barcodes.
type Code39 struct{ Barcode1DBase }

// NewCode39 creates a Code39 barcode generator.
func NewCode39(outputFormat string) *Code39 {
	bc := &Code39{Barcode1DBase: new1DBase("Code39")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetShowStartStop sets whether to show start/stop characters.
func (b *Code39) SetShowStartStop(show bool) { b.settings["showStartStop"] = show }

// Code93 generates Code93 barcodes.
type Code93 struct{ Barcode1DBase }

// NewCode93 creates a Code93 barcode generator.
func NewCode93(outputFormat string) *Code93 {
	bc := &Code93{Barcode1DBase: new1DBase("Code93")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// Code128 generates Code128 barcodes.
type Code128 struct{ Barcode1DBase }

// NewCode128 creates a Code128 barcode generator.
func NewCode128(outputFormat string) *Code128 {
	bc := &Code128{Barcode1DBase: new1DBase("Code128")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetCodeMode sets the code mode (AUTO, A, B, C).
func (b *Code128) SetCodeMode(mode string) { b.settings["codeMode"] = mode }

// GS1128 generates GS1-128 barcodes.
type GS1128 struct{ Barcode1DBase }

// NewGS1128 creates a GS1-128 barcode generator.
func NewGS1128(outputFormat string) *GS1128 {
	bc := &GS1128{Barcode1DBase: new1DBase("GS1_128")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// NW7 generates NW-7 (Codabar) barcodes.
type NW7 struct{ Barcode1DBase }

// NewNW7 creates a NW-7 barcode generator.
func NewNW7(outputFormat string) *NW7 {
	bc := &NW7{Barcode1DBase: new1DBase("NW7")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetShowStartStop sets whether to show start/stop characters.
func (b *NW7) SetShowStartStop(show bool) { b.settings["showStartStop"] = show }

// ITF generates ITF (Interleaved 2 of 5) barcodes.
type ITF struct{ Barcode1DBase }

// NewITF creates an ITF barcode generator.
func NewITF(outputFormat string) *ITF {
	bc := &ITF{Barcode1DBase: new1DBase("ITF")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// Matrix2of5 generates Matrix 2 of 5 barcodes.
type Matrix2of5 struct{ Barcode1DBase }

// NewMatrix2of5 creates a Matrix 2 of 5 barcode generator.
func NewMatrix2of5(outputFormat string) *Matrix2of5 {
	bc := &Matrix2of5{Barcode1DBase: new1DBase("Matrix2of5")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// NEC2of5 generates NEC 2 of 5 barcodes.
type NEC2of5 struct{ Barcode1DBase }

// NewNEC2of5 creates a NEC 2 of 5 barcode generator.
func NewNEC2of5(outputFormat string) *NEC2of5 {
	bc := &NEC2of5{Barcode1DBase: new1DBase("NEC2of5")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// Jan8 generates JAN-8 (EAN-8) barcodes.
type Jan8 struct{ Barcode1DBase }

// NewJAN8 creates a JAN-8 barcode generator.
func NewJAN8(outputFormat string) *Jan8 {
	bc := &Jan8{Barcode1DBase: new1DBase("Jan8")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetExtendedGuard sets whether to use extended guard bars.
func (b *Jan8) SetExtendedGuard(ext bool) { b.settings["extendedGuard"] = ext }

// Jan13 generates JAN-13 (EAN-13) barcodes.
type Jan13 struct{ Barcode1DBase }

// NewJAN13 creates a JAN-13 barcode generator.
func NewJAN13(outputFormat string) *Jan13 {
	bc := &Jan13{Barcode1DBase: new1DBase("Jan13")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetExtendedGuard sets whether to use extended guard bars.
func (b *Jan13) SetExtendedGuard(ext bool) { b.settings["extendedGuard"] = ext }

// UPCA generates UPC-A barcodes.
type UPCA struct{ Barcode1DBase }

// NewUPCA creates a UPC-A barcode generator.
func NewUPCA(outputFormat string) *UPCA {
	bc := &UPCA{Barcode1DBase: new1DBase("UPC_A")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetExtendedGuard sets whether to use extended guard bars.
func (b *UPCA) SetExtendedGuard(ext bool) { b.settings["extendedGuard"] = ext }

// UPCE generates UPC-E barcodes.
type UPCE struct{ Barcode1DBase }

// NewUPCE creates a UPC-E barcode generator.
func NewUPCE(outputFormat string) *UPCE {
	bc := &UPCE{Barcode1DBase: new1DBase("UPC_E")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetExtendedGuard sets whether to use extended guard bars.
func (b *UPCE) SetExtendedGuard(ext bool) { b.settings["extendedGuard"] = ext }

// ═════════════════════════════════════════════════════════════════════════════
// GS1 DataBar
// ═════════════════════════════════════════════════════════════════════════════

// GS1DataBar14 generates GS1 DataBar 14 barcodes.
type GS1DataBar14 struct{ Barcode1DBase }

// NewGS1DataBar14 creates a GS1 DataBar 14 barcode generator.
func NewGS1DataBar14(outputFormat string) *GS1DataBar14 {
	bc := &GS1DataBar14{Barcode1DBase: new1DBase("GS1DataBar14")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetSymbolType sets the symbol type (OMNIDIRECTIONAL, STACKED, STACKED_OMNIDIRECTIONAL).
func (b *GS1DataBar14) SetSymbolType(symbolType string) { b.settings["symbolType"] = symbolType }

// GS1DataBarLimited generates GS1 DataBar Limited barcodes.
type GS1DataBarLimited struct{ Barcode1DBase }

// NewGS1DataBarLimited creates a GS1 DataBar Limited barcode generator.
func NewGS1DataBarLimited(outputFormat string) *GS1DataBarLimited {
	bc := &GS1DataBarLimited{Barcode1DBase: new1DBase("GS1DataBarLimited")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// GS1DataBarExpanded generates GS1 DataBar Expanded barcodes.
type GS1DataBarExpanded struct{ Barcode1DBase }

// NewGS1DataBarExpanded creates a GS1 DataBar Expanded barcode generator.
func NewGS1DataBarExpanded(outputFormat string) *GS1DataBarExpanded {
	bc := &GS1DataBarExpanded{Barcode1DBase: new1DBase("GS1DataBarExpanded")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetSymbolType sets the symbol type (UNSTACKED, STACKED).
func (b *GS1DataBarExpanded) SetSymbolType(symbolType string) {
	b.settings["symbolType"] = symbolType
}

// SetNoOfColumns sets the number of columns for stacked version.
func (b *GS1DataBarExpanded) SetNoOfColumns(cols int) { b.settings["noOfColumns"] = cols }

// ═════════════════════════════════════════════════════════════════════════════
// Special Barcodes
// ═════════════════════════════════════════════════════════════════════════════

// YubinCustomer generates Japanese postal customer barcodes.
type YubinCustomer struct {
	BarcodeWasmBase
}

// NewYubinCustomer creates a YubinCustomer barcode generator.
func NewYubinCustomer(outputFormat string) *YubinCustomer {
	bc := &YubinCustomer{BarcodeWasmBase: newBase("YubinCustomer")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetPxAdjustBlack sets pixel adjustment for black bars.
func (b *YubinCustomer) SetPxAdjustBlack(adj int) { b.settings["pxAdjustBlack"] = adj }

// SetPxAdjustWhite sets pixel adjustment for white bars.
func (b *YubinCustomer) SetPxAdjustWhite(adj int) { b.settings["pxAdjustWhite"] = adj }

// Draw generates a postal barcode. Width is auto-calculated.
func (b *YubinCustomer) Draw(code string, height int) (string, error) {
	return b.call("draw", code, height)
}

// DrawWithWidth generates a postal barcode with explicit width.
func (b *YubinCustomer) DrawWithWidth(code string, width, height int) (string, error) {
	return b.call("drawWithWidth", code, width, height)
}

// ═════════════════════════════════════════════════════════════════════════════
// 2D Barcodes
// ═════════════════════════════════════════════════════════════════════════════

// QR generates QR codes.
type QR struct{ Barcode2DBase }

// NewQRCode creates a QR code generator.
func NewQRCode(outputFormat string) *QR {
	bc := &QR{Barcode2DBase: new2DBase("QR")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetErrorCorrectionLevel sets the error correction level (L, M, Q, H).
func (b *QR) SetErrorCorrectionLevel(level string) { b.settings["errorCorrectionLevel"] = level }

// SetVersion sets QR version (0=auto, 1-40).
func (b *QR) SetVersion(version int) { b.settings["version"] = version }

// SetEncodeMode sets the encode mode (NUMERIC, ALPHANUMERIC, BYTE, KANJI).
func (b *QR) SetEncodeMode(mode string) { b.settings["encodeMode"] = mode }

// DataMatrix generates DataMatrix barcodes.
type DataMatrix struct{ Barcode2DBase }

// NewDataMatrix creates a DataMatrix barcode generator.
func NewDataMatrix(outputFormat string) *DataMatrix {
	bc := &DataMatrix{Barcode2DBase: new2DBase("DataMatrix")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetCodeSize sets the code size (AUTO, 10x10, 12x12, etc.).
func (b *DataMatrix) SetCodeSize(size string) { b.settings["codeSize"] = size }

// SetEncodeScheme sets the encode scheme (AUTO, ASCII, C40, TEXT, X12, EDIFACT, BASE256).
func (b *DataMatrix) SetEncodeScheme(scheme string) { b.settings["encodeScheme"] = scheme }

// PDF417 generates PDF417 barcodes.
type PDF417 struct{ Barcode2DBase }

// NewPDF417 creates a PDF417 barcode generator.
func NewPDF417(outputFormat string) *PDF417 {
	bc := &PDF417{Barcode2DBase: new2DBase("PDF417")}
	bc.SetOutputFormat(outputFormat)
	return bc
}

// SetErrorLevel sets the error correction level (-1=auto, 0-8).
func (b *PDF417) SetErrorLevel(level int) { b.settings["errorLevel"] = level }

// SetColumns sets the number of columns.
func (b *PDF417) SetColumns(cols int) { b.settings["columns"] = cols }

// SetRows sets the number of rows.
func (b *PDF417) SetRows(rows int) { b.settings["rows"] = rows }

// SetAspectRatio sets the aspect ratio.
func (b *PDF417) SetAspectRatio(ratio float64) { b.settings["aspectRatio"] = ratio }

// SetYHeight sets the Y height.
func (b *PDF417) SetYHeight(yHeight int) { b.settings["yHeight"] = yHeight }

// Draw generates a PDF417 barcode (width × height).
func (b *PDF417) Draw(code string, width, height int) (string, error) {
	return b.call("draw", code, width, height)
}

// ═════════════════════════════════════════════════════════════════════════════
// Product Info
// ═════════════════════════════════════════════════════════════════════════════

// GetProductName returns the product name.
func GetProductName() string { return "barcode-pao-wasm (Go)" }

// GetVersion returns the version.
func GetVersion() string { return "1.2.0" }

// GetManufacturer returns the manufacturer.
func GetManufacturer() string { return "Pao" }
