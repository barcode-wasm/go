
import Module from './barcode.mjs';

const args = JSON.parse(process.argv[2]);
const {className, method, methodArgs, settings} = args;

try {
    const m = await Module();
    const BarcodeClass = m[className];
    if (!BarcodeClass) {
        console.log(JSON.stringify({error: `Unknown class: ${className}`}));
        process.exit(1);
    }

    const instance = new BarcodeClass();

    // Apply settings
    if (settings.outputFormat) instance.setOutputFormat(settings.outputFormat);
    if (settings.foregroundColor) {
        const [r, g, b, a] = settings.foregroundColor;
        instance.setForegroundColor(r, g, b, a);
    }
    if (settings.backgroundColor) {
        const [r, g, b, a] = settings.backgroundColor;
        instance.setBackgroundColor(r, g, b, a);
    }
    if (settings.showText !== undefined && instance.setShowText) instance.setShowText(settings.showText);
    if (settings.fitWidth !== undefined && instance.setFitWidth) instance.setFitWidth(settings.fitWidth);
    if (settings.pxAdjustBlack !== undefined && instance.setPxAdjustBlack) instance.setPxAdjustBlack(settings.pxAdjustBlack);
    if (settings.pxAdjustWhite !== undefined && instance.setPxAdjustWhite) instance.setPxAdjustWhite(settings.pxAdjustWhite);
    if (settings.textGap !== undefined && instance.setTextGap) instance.setTextGap(settings.textGap);
    if (settings.textFontScale !== undefined && instance.setTextFontScale) instance.setTextFontScale(settings.textFontScale);
    if (settings.textEvenSpacing !== undefined && instance.setTextEvenSpacing) instance.setTextEvenSpacing(settings.textEvenSpacing);
    if (settings.showStartStop !== undefined && instance.setShowStartStop) instance.setShowStartStop(settings.showStartStop);
    if (settings.codeMode !== undefined && instance.setCodeMode) instance.setCodeMode(settings.codeMode);
    if (settings.extendedGuard !== undefined && instance.setExtendedGuard) instance.setExtendedGuard(settings.extendedGuard);
    if (settings.symbolType !== undefined && instance.setSymbolType) instance.setSymbolType(settings.symbolType);
    if (settings.noOfColumns !== undefined && instance.setNoOfColumns) instance.setNoOfColumns(settings.noOfColumns);
    if (settings.stringEncoding !== undefined && instance.setStringEncoding) instance.setStringEncoding(settings.stringEncoding);
    if (settings.errorCorrectionLevel !== undefined && instance.setErrorCorrectionLevel) instance.setErrorCorrectionLevel(settings.errorCorrectionLevel);
    if (settings.version !== undefined && instance.setVersion) instance.setVersion(settings.version);
    if (settings.encodeMode !== undefined && instance.setEncodeMode) instance.setEncodeMode(settings.encodeMode);
    if (settings.codeSize !== undefined && instance.setCodeSize) instance.setCodeSize(settings.codeSize);
    if (settings.encodeScheme !== undefined && instance.setEncodeScheme) instance.setEncodeScheme(settings.encodeScheme);
    if (settings.errorLevel !== undefined && instance.setErrorLevel) instance.setErrorLevel(settings.errorLevel);
    if (settings.columns !== undefined && instance.setColumns) instance.setColumns(settings.columns);
    if (settings.rows !== undefined && instance.setRows) instance.setRows(settings.rows);
    if (settings.aspectRatio !== undefined && instance.setAspectRatio) instance.setAspectRatio(settings.aspectRatio);
    if (settings.yHeight !== undefined && instance.setYHeight) instance.setYHeight(settings.yHeight);

    // Call the method
    const func = instance[method];
    if (!func) {
        console.log(JSON.stringify({error: `Unknown method: ${method}`}));
        process.exit(1);
    }

    const result = func.apply(instance, methodArgs);
    console.log(JSON.stringify({result: result}));
} catch (err) {
    console.log(JSON.stringify({error: err.message || String(err)}));
    process.exit(1);
}
