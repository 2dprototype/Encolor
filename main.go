package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color mapping for 16-bit (hexadecimal) characters
var colorMap16bit = map[byte][3]uint8{
	'0': {0, 0, 0},       // Black
	'1': {255, 255, 255}, // White
	'2': {255, 0, 0},     // Red
	'3': {0, 255, 0},     // Green
	'4': {0, 0, 255},     // Blue
	'5': {255, 255, 0},   // Yellow
	'6': {255, 0, 255},   // Magenta
	'7': {0, 255, 255},   // Cyan
	'8': {128, 128, 128}, // Gray
	'9': {192, 192, 192}, // Light Gray

	'a': {128, 0, 0},   // Dark Red
	'b': {0, 128, 0},   // Dark Green
	'c': {0, 0, 128},   // Dark Blue
	'd': {128, 128, 0}, // Olive
	'e': {128, 0, 128}, // Purple
	'f': {0, 128, 128}, // Teal
}

// Color mapping for 8-bit (octal) characters
var colorMap8bit = map[byte][3]uint8{
	'0': {0, 0, 0},       // Black
	'1': {255, 255, 255}, // White
	'2': {255, 0, 0},     // Red
	'3': {0, 255, 0},     // Green
	'4': {0, 0, 255},     // Blue
	'5': {255, 255, 0},   // Yellow
	'6': {255, 0, 255},   // Magenta
	'7': {0, 255, 255},   // Cyan
}

// Reverse color mapping for reconstruction
var reverseColorMap16bit = createReverseColorMap(colorMap16bit)
var reverseColorMap8bit = createReverseColorMap(colorMap8bit)

func main() {
	// Define command line flags
	inputFile := flag.String("i", "", "Input file to convert")
	outputFile := flag.String("o", "", "Output image file (PNG)")
	colorMode := flag.String("m", "16bit", "Color mode: 8bit or 16bit")
	reconstruct := flag.Bool("re", false, "Reconstruct file from image")
	schemeFile := flag.String("sch", "", "Color scheme file (or name from scheme folder)")
	listSchemes := flag.Bool("sch-list", false, "List all available color schemes")
	flag.Parse()

	// Handle scheme listing
	if *listSchemes {
		listAvailableSchemes()
		return
	}

	// Load color schemes
	var err error
	if *schemeFile != "" {
		// Load custom scheme
		err = loadCustomScheme(*schemeFile)
		if err != nil {
			log.Printf("Warning: Could not load custom scheme: %v", err)
		} else {
			fmt.Printf("Loaded custom scheme: %s\n", *schemeFile)
		}
	} else {
		// Load default scheme
		err = loadDefaultScheme()
		if err != nil {
			log.Printf("Warning: Could not load default scheme: %v", err)
		}
	}

	// Validate input
	if *inputFile == "" && !*reconstruct {
		fmt.Println("Usage:")
		fmt.Println("  Encode: colorcode -i <inputfile> [-o <output.png>] [-m 8bit|16bit] [-sch <scheme>]")
		fmt.Println("  Decode: colorcode -re -i <image.png> [-o <outputfile>]")
		fmt.Println("  List schemes: colorcode -sch-list")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *reconstruct {
		// Reconstruct file from image
		err := reconstructFileFromImage(*inputFile, *outputFile)
		if err != nil {
			log.Fatalf("Error reconstructing file: %v", err)
		}
		fmt.Printf("Successfully reconstructed file: %s\n", *outputFile)
	} else {
		// Encode file to image
		err := encodeFileToImage(*inputFile, *outputFile, *colorMode)
		if err != nil {
			log.Fatalf("Error encoding file: %v", err)
		}
	}
}

func loadDefaultScheme() error {
	// Get the directory of the executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	schemeFile := filepath.Join(exeDir, "scheme.ini")

	// Check if scheme.ini exists
	if _, err := os.Stat(schemeFile); os.IsNotExist(err) {
		// Create default scheme.ini
		err := createDefaultSchemeFile(schemeFile)
		if err != nil {
			return fmt.Errorf("error creating default scheme file: %v", err)
		}
		fmt.Printf("Created default scheme file: %s\n", schemeFile)
		return nil
	}

	// Load scheme.ini
	err = parseSchemeFile(schemeFile)
	if err != nil {
		return fmt.Errorf("error parsing scheme file: %v", err)
	}

	fmt.Printf("Loaded default scheme from: %s\n", schemeFile)
	return nil
}

func loadCustomScheme(schemeName string) error {
	// Get the directory of the executable
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("error getting executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	
	var schemeFile string
	
	// Check if schemeName is a full path
	if strings.Contains(schemeName, string(filepath.Separator)) || strings.HasSuffix(schemeName, ".ini") {
		// Use as direct file path
		schemeFile = schemeName
	} else {
		// Look in scheme directory
		schemeDir := filepath.Join(exeDir, "scheme")
		schemeFile = filepath.Join(schemeDir, schemeName+".ini")
	}
	
	// Check if file exists
	if _, err := os.Stat(schemeFile); os.IsNotExist(err) {
		return fmt.Errorf("scheme file not found: %s", schemeFile)
	}
	
	// Load the scheme file
	err = parseSchemeFile(schemeFile)
	if err != nil {
		return fmt.Errorf("error parsing scheme file: %v", err)
	}
	
	return nil
}

func listAvailableSchemes() {
	// Get the directory of the executable
	exePath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		return
	}
	exeDir := filepath.Dir(exePath)
	schemeDir := filepath.Join(exeDir, "scheme")
	
	// Check if scheme directory exists
	if _, err := os.Stat(schemeDir); os.IsNotExist(err) {
		fmt.Printf("Scheme directory not found: %s\n", schemeDir)
		fmt.Println("No pre-packed schemes available.")
		return
	}
	
	// Read scheme directory
	files, err := os.ReadDir(schemeDir)
	if err != nil {
		fmt.Printf("Error reading scheme directory: %v\n", err)
		return
	}
	
	fmt.Println("Available Color Schemes")
	fmt.Println("=========================")
	
	var schemeFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".ini") {
			schemeFiles = append(schemeFiles, file.Name())
		}
	}
	
	if len(schemeFiles) == 0 {
		fmt.Println("No scheme files found in scheme directory.")
		return
	}
	
	// Display schemes with color previews
	for i, schemeFile := range schemeFiles {
		schemeName := strings.TrimSuffix(schemeFile, ".ini")
		fullPath := filepath.Join(schemeDir, schemeFile)
		
		fmt.Printf("\n%2d. %s\n", i+1, schemeName)
		
		// Load and display color preview
		preview, err := generateSchemePreview(fullPath)
		if err != nil {
			fmt.Printf("   Error loading scheme: %v\n", err)
			continue
		}
		
		fmt.Println(preview)
	}
	
	fmt.Println("\nUsage: -sch <scheme_name>")
	fmt.Println("   Example: colorcode -i file.txt -sch monokai")
}

func generateSchemePreview(schemeFile string) (string, error) {
	// Parse the scheme file to get colors
	_, temp16bit, err := parseSchemeFileForPreview(schemeFile)
	if err != nil {
		return "", err
	}
	
	var colorBlocks []string
	
	// Create color blocks for 16-bit scheme
	if len(temp16bit) > 0 {
		blocks16bit := createColorBlocks(temp16bit)
		colorBlocks = append(colorBlocks, blocks16bit)
	}
	
	return lipgloss.JoinVertical(lipgloss.Left, colorBlocks...), nil
}

func parseSchemeFileForPreview(filename string) (map[byte][3]uint8, map[byte][3]uint8, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	temp8bit := make(map[byte][3]uint8)
	temp16bit := make(map[byte][3]uint8)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		// Parse color mappings
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}

		char := parts[0]
		if len(char) != 1 {
			continue
		}

		rgbParts := strings.Split(parts[1], ",")
		if len(rgbParts) != 3 {
			continue
		}

		r, err1 := strconv.Atoi(rgbParts[0])
		g, err2 := strconv.Atoi(rgbParts[1])
		b, err3 := strconv.Atoi(rgbParts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		// Ensure RGB values are in valid range
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			continue
		}

		colorValue := [3]uint8{uint8(r), uint8(g), uint8(b)}

		// Add to appropriate map
		switch currentSection {
		case "8bit":
			temp8bit[char[0]] = colorValue
		case "16bit":
			temp16bit[char[0]] = colorValue
		}
	}

	return temp8bit, temp16bit, scanner.Err()
}

func createColorBlocks(colorMap map[byte][3]uint8) string {
	var blocks []string
	var colorLines []string
	
	// Sort characters for consistent display
	chars := make([]byte, 0, len(colorMap))
	for char := range colorMap {
		chars = append(chars, char)
	}
	
	// Sort: numbers first, then letters
	for i := 0; i < len(chars); i++ {
		for j := i + 1; j < len(chars); j++ {
			if chars[i] > chars[j] {
				chars[i], chars[j] = chars[j], chars[i]
			}
		}
	}
	
	// Create color blocks
	for _, char := range chars {
		rgb := colorMap[char]
		colorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2]))).
			Foreground(getContrastColor(rgb)).
			Width(4).
			Height(1).
			Align(lipgloss.Center).
			Bold(true)
		
		block := colorStyle.Render(string(char))
		blocks = append(blocks, block)
	}
	
	// Group blocks into lines of 8
	for i := 0; i < len(blocks); i += 8 {
		end := i + 8
		if end > len(blocks) {
			end = len(blocks)
		}
		line := lipgloss.JoinHorizontal(lipgloss.Left, blocks[i:end]...)
		colorLines = append(colorLines, "   "+line)
	}

	// Combine title and color blocks
	result := []string{}
	result = append(result, colorLines...)
	
	return lipgloss.JoinVertical(lipgloss.Left, result...)
}

func getContrastColor(rgb [3]uint8) lipgloss.Color {
	// Calculate relative luminance (simplified)
	luminance := (0.299*float64(rgb[0]) + 0.587*float64(rgb[1]) + 0.114*float64(rgb[2])) / 255
	
	if luminance > 0.5 {
		return lipgloss.Color("#000000") // Use black text on light backgrounds
	}
	return lipgloss.Color("#ffffff") // Use white text on dark backgrounds
}

func createDefaultSchemeFile(filename string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write header
	writer.WriteString("# Color Scheme Configuration\n")
	writer.WriteString("# Format: character=red,green,blue\n")
	writer.WriteString("# RGB values range from 0 to 255\n\n")

	// Write 8-bit scheme
	writer.WriteString("[8bit]\n")
	for char, rgb := range colorMap8bit {
		writer.WriteString(fmt.Sprintf("%c=%d,%d,%d\n", char, rgb[0], rgb[1], rgb[2]))
	}

	writer.WriteString("\n")

	// Write 16-bit scheme
	writer.WriteString("[16bit]\n")
	for char, rgb := range colorMap16bit {
		writer.WriteString(fmt.Sprintf("%c=%d,%d,%d\n", char, rgb[0], rgb[1], rgb[2]))
	}

	return writer.Flush()
}

func parseSchemeFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	currentSection := ""

	// Temporary maps to store loaded schemes
	temp8bit := make(map[byte][3]uint8)
	temp16bit := make(map[byte][3]uint8)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentSection = line[1 : len(line)-1]
			continue
		}

		// Parse color mappings
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue // Skip invalid lines
		}

		char := parts[0]
		if len(char) != 1 {
			continue // Skip if not a single character
		}

		rgbParts := strings.Split(parts[1], ",")
		if len(rgbParts) != 3 {
			continue // Skip if not exactly 3 RGB values
		}

		r, err1 := strconv.Atoi(rgbParts[0])
		g, err2 := strconv.Atoi(rgbParts[1])
		b, err3 := strconv.Atoi(rgbParts[2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue // Skip if RGB values are invalid
		}

		// Ensure RGB values are in valid range
		if r < 0 || r > 255 || g < 0 || g > 255 || b < 0 || b > 255 {
			continue
		}

		colorValue := [3]uint8{uint8(r), uint8(g), uint8(b)}

		// Add to appropriate map based on current section
		switch currentSection {
		case "8bit":
			temp8bit[char[0]] = colorValue
		case "16bit":
			temp16bit[char[0]] = colorValue
		}
	}

	// Update global color maps if we found valid entries
	if len(temp8bit) > 0 {
		colorMap8bit = temp8bit
		reverseColorMap8bit = createReverseColorMap(colorMap8bit)
	}

	if len(temp16bit) > 0 {
		colorMap16bit = temp16bit
		reverseColorMap16bit = createReverseColorMap(colorMap16bit)
	}

	return scanner.Err()
}

func encodeFileToImage(inputFile, outputFile, colorMode string) error {
	// Read the input file
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("error reading input file: %v", err)
	}

	fmt.Printf("Read %d bytes from %s\n", len(data), inputFile)

	var encodedData string
	var colorMap map[byte][3]uint8

	if colorMode == "8bit" {
		// Convert data to octal (base8)
		encodedData = convertToOctal(data)
		colorMap = colorMap8bit
		fmt.Printf("Converted to %d octal characters (8-bit mode)\n", len(encodedData))
	} else {
		// Convert data to hexadecimal (base16)
		encodedData = hex.EncodeToString(data)
		colorMap = colorMap16bit
		fmt.Printf("Converted to %d hex characters (16-bit mode)\n", len(encodedData))
	}

	// Create image from encoded data
	img, err := createImageFromEncodedData(encodedData, colorMap, colorMode)
	if err != nil {
		return fmt.Errorf("error creating image: %v", err)
	}

	// Save the image
	if outputFile == "" {
		outputFile = inputFile + "_encoded.png"
	}

	err = saveImage(img, outputFile)
	if err != nil {
		return fmt.Errorf("error saving image: %v", err)
	}

	fmt.Printf("Successfully created image: %s\n", outputFile)
	fmt.Printf("Image dimensions: %d x %d pixels\n", img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Printf("Color mode: %s\n", colorMode)
	return nil
}

func reconstructFileFromImage(inputImage, outputFile string) error {
	// Read the image file
	file, err := os.Open(inputImage)
	if err != nil {
		return fmt.Errorf("error opening image file: %v", err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		return fmt.Errorf("error decoding image: %v", err)
	}

	fmt.Printf("Read image with dimensions: %d x %d\n", img.Bounds().Dx(), img.Bounds().Dy())

	// Reconstruct data from image
	reconstructedData, colorMode, err := reconstructDataFromImage(img)
	if err != nil {
		return fmt.Errorf("error reconstructing data: %v", err)
	}

	fmt.Printf("Reconstructed data using %s color mode\n", colorMode)

	// Convert back to binary data
	var binaryData []byte
	if colorMode == "8bit" {
		binaryData, err = convertFromOctal(reconstructedData)
		if err != nil {
			return fmt.Errorf("error converting from octal: %v", err)
		}
	} else {
		binaryData, err = hex.DecodeString(reconstructedData)
		if err != nil {
			return fmt.Errorf("error decoding hex: %v", err)
		}
	}

	// Save the reconstructed file
	if outputFile == "" {
		outputFile = inputImage + "_decoded"
	}
	err = os.WriteFile(outputFile, binaryData, 0644)
	if err != nil {
		return fmt.Errorf("error writing reconstructed file: %v", err)
	}

	fmt.Printf("Reconstructed %d bytes of data\n", len(binaryData))
	return nil
}

func createImageFromEncodedData(encodedData string, colorMap map[byte][3]uint8, colorMode string) (*image.RGBA, error) {
	// Calculate image dimensions to be as square as possible
	dataLength := len(encodedData)
	width := int(math.Ceil(math.Sqrt(float64(dataLength))))
	height := int(math.Ceil(float64(dataLength) / float64(width)))

	fmt.Printf("Creating image with dimensions: %d x %d\n", width, height)

	// Create a new RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Set pixels based on encoded data
	for i, char := range encodedData {
		if i >= width*height {
			break
		}

		x := i % width
		y := i / width

		// Get color for this character
		colorVal, exists := colorMap[byte(char)]
		if !exists {
			// Use black for unknown characters
			colorVal = [3]uint8{0, 0, 0}
		}

		// Set the pixel color
		img.Set(x, y, color.RGBA{
			R: colorVal[0],
			G: colorVal[1],
			B: colorVal[2],
			A: 255,
		})
	}

	// Fill remaining pixels with white (to distinguish from data)
	for i := len(encodedData); i < width*height; i++ {
		x := i % width
		y := i / width
		img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	}

	return img, nil
}

func reconstructDataFromImage(img image.Image) (string, string, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	whitePixels := 0

	// Try both color modes and see which one gives more valid characters
	var results16bit, results8bit strings.Builder
	valid16bit := 0
	valid8bit := 0

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := img.At(x, y)
			r, g, b, _ := pixel.RGBA()
			// Convert from 16-bit to 8-bit color
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Check if pixel is white (end of data)
			if r8 == 255 && g8 == 255 && b8 == 255 {
				whitePixels++
				continue
			}

			color := [3]uint8{r8, g8, b8}

			// Try 16-bit mapping
			if char, exists := reverseColorMap16bit[color]; exists {
				results16bit.WriteByte(char)
				valid16bit++
			}

			// Try 8-bit mapping
			if char, exists := reverseColorMap8bit[color]; exists {
				results8bit.WriteByte(char)
				valid8bit++
			}
		}
	}

	// Determine which color mode was used based on valid character count
	if valid16bit >= valid8bit {
		fmt.Printf("Detected 16-bit color mode (%d valid characters)\n", valid16bit)
		return results16bit.String(), "16bit", nil
	} else {
		fmt.Printf("Detected 8-bit color mode (%d valid characters)\n", valid8bit)
		return results8bit.String(), "8bit", nil
	}
}

func createReverseColorMap(colorMap map[byte][3]uint8) map[[3]uint8]byte {
	reverseMap := make(map[[3]uint8]byte)
	for char, color := range colorMap {
		reverseMap[color] = char
	}
	return reverseMap
}

func convertToOctal(data []byte) string {
	var result strings.Builder
	for _, b := range data {
		// Convert each byte to 3-digit octal
		result.WriteString(fmt.Sprintf("%03o", b))
	}
	return result.String()
}

func convertFromOctal(octalStr string) ([]byte, error) {
	if len(octalStr)%3 != 0 {
		return nil, fmt.Errorf("octal string length must be multiple of 3")
	}

	result := make([]byte, len(octalStr)/3)
	for i := 0; i < len(octalStr); i += 3 {
		octalByte := octalStr[i : i+3]
		value, err := strconv.ParseUint(octalByte, 8, 8)
		if err != nil {
			return nil, fmt.Errorf("invalid octal digit: %s", octalByte)
		}
		result[i/3] = byte(value)
	}
	return result, nil
}

func saveImage(img *image.RGBA, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return png.Encode(file, img)
}