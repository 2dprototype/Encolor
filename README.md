# Encolor

> Encode files as colorful PNG images and decode them back!

# Preview
Here is a encoded image of a rar file

![rar](./preview/rar.png)

Also, check `preview` folder for more images.

## Features
- **Encode** any file into a colorful PNG image
- **Decode** images back to original files  
- **40+ color schemes** included (modern, monokai, dracula, kawaii, military, etc.)
- **Visual scheme previews** with `-sch-list`
- **Custom schemes** support via INI files

## Quick Start

### Encode a File
```bash
# Basic encoding (uses default scheme)
encolor -i secret.txt -o encoded.png

# With custom color scheme
encolor -i secret.txt -o encoded.png -sch monokai

# 8-bit mode (smaller images)
encolor -i secret.txt -o encoded.png -m 8bit -sch demon
```

### Decode an Image
```bash
encolor -re -i encoded.png -o decoded.txt
```

### Browse Schemes
```bash
encolor -sch-list
```

## Color Schemes

Use `-sch-list` to see all 68+ schemes with visual previews:

```bash
# Some popular schemes:
-sch modern      # Sleek modern colors
-sch monokai     # Popular code theme  
-sch dracula     # Dark theme
-sch demon       # Hellfire reds
-sch military    # Camo greens
-sch rainbow     # Vibrant colors
-sch kawaii      # Cute pastels
-sch cyberpunk   # Neon futurism
```

## Scheme Files

Create custom schemes in `scheme/yourtheme.ini`:
```ini
[8bit]
0=0,0,0
1=255,255,255
2=255,0,0
# ... etc

[16bit]  
0=0,0,0
1=255,255,255
2=255,0,0
# ... etc up to 'f'
```

## Installation

1. Download the binary for your OS
2. Run directly - no installation needed!

## How It Works

- **16-bit mode**: Files → Hex → Colors → PNG
- **8-bit mode**: Files → Octal → Colors → PNG  
- Each character = 1 colored pixel
- White pixels mark end of data

## Usage
```
encolor -i <input> [-o <output.png>] [-m 8bit|16bit] [-sch <scheme>]
encolor -re -i <image.png> [-o <output>]
encolor -sch-list
```