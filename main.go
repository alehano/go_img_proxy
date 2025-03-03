package main

import (
	"log"
	"net/http"
	"os"

	"io"
	"strconv"
	"strings"

	"image"

	"image/jpeg"
	"image/png"

	"fmt"

	"net/url"

	"crypto/tls"

	"encoding/base64"
	"encoding/json"

	"github.com/anthonynsimon/bild/adjust"
	"github.com/anthonynsimon/bild/effect"
	"github.com/jessevdk/go-flags"
	"github.com/sunshineplan/imgconv"
)

// Define custom position constants
const (
	TopLeft     = "topleft"
	TopRight    = "topright"
	BottomLeft  = "bottomleft"
	BottomRight = "bottomright"
	Center      = "center"
)

type Config struct {
	Debug                bool   `long:"debug" env:"DEBUG" description:"Enable debug mode"`
	Quality              int    `long:"quality" env:"QUALITY" default:"85" description:"Quality of the JPEG image"`
	Port                 int    `long:"port" env:"PORT" default:"8080" description:"Port to run the server on"`
	WatermarksConfigFile string `long:"watermarks-config" env:"WATERMARKS_CONFIG_FILE" default:"watermarks.json" description:"Path to the watermarks config file"`
}

// New struct for WatermarkConfig
// This struct will hold the watermark configuration for each project

type WatermarkConfig struct {
	Image          string
	Opacity        int
	SizePercent    int
	OffsetXPercent int
	OffsetYPercent int
	Position       string
	Force          bool
	Random         bool
}

func parseConfig() (*Config, error) {
	var opts Config
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			p.WriteHelp(os.Stderr)
		}
		return nil, err
	}
	return &opts, nil
}

// Calculate the watermark offset relative to the centered position.
// The desired absolute position is computed based on the provided position and margin percentages.
func calculateWatermarkOffset(position string, imgWidth, imgHeight, watermarkWidth, watermarkHeight int, offsetXPercent, offsetYPercent int) (offsetX, offsetY int) {
	// Calculate the center position (where the library would place the watermark if no offset is given).
	centerX := (imgWidth - watermarkWidth) / 2
	centerY := (imgHeight - watermarkHeight) / 2

	// Calculate margin (in pixels) from the corresponding image border.
	marginX := imgWidth * offsetXPercent / 100
	marginY := imgHeight * offsetYPercent / 100

	// Determine the desired absolute watermark position.
	var absX, absY int
	switch strings.ToLower(position) {
	case TopLeft:
		absX = marginX
		absY = marginY
	case TopRight:
		absX = imgWidth - watermarkWidth - marginX
		absY = marginY
	case BottomLeft:
		absX = marginX
		absY = imgHeight - watermarkHeight - marginY
	case BottomRight:
		absX = imgWidth - watermarkWidth - marginX
		absY = imgHeight - watermarkHeight - marginY
	default: // Center (or any unrecognized value defaults to center)
		absX = centerX
		absY = centerY
	}

	// Return the offset relative to the center position.
	return absX - centerX, absY - centerY
}

func parseOptions(options string) (map[string]string, error) {
	// Split the options string by underscores
	params := strings.Split(options, "_")
	optionsMap := make(map[string]string)
	for _, param := range params {
		// Split each parameter by the first hyphen
		kv := strings.SplitN(param, "-", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid option format: %s", param)
		}
		optionsMap[kv[0]] = kv[1]
	}
	return optionsMap, nil
}

// Function to load watermark configurations from a JSON file
func loadWatermarkConfigs(filePath string) (map[string]WatermarkConfig, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var configs map[string]WatermarkConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configs)
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func processImage(w http.ResponseWriter, r *http.Request, cfg *Config, watermarkConfigs map[string]WatermarkConfig) {
	// Extract the path and split it to get options and URL
	pathParts := strings.SplitN(r.URL.Path[1:], "/", 3)
	if len(pathParts) != 3 {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	options := pathParts[0]
	urlType := pathParts[1]
	rawURL := pathParts[2]

	// Parse options
	optionsMap, err := parseOptions(options)
	if err != nil {
		http.Error(w, "Invalid options format", http.StatusBadRequest)
		return
	}

	// Get the project from options, default to "default"
	project := optionsMap["p"]
	if project == "" {
		project = "default"
	}

	// Get the watermark configuration for the project
	watermarkConfig, exists := watermarkConfigs[project]
	if !exists {
		watermarkConfig = watermarkConfigs["default"]
	}

	var imageURL string
	if urlType == "url" {
		imageURL = rawURL
	} else if urlType == "urlb" {
		// Trim suffix after a dot before decoding
		if dotIndex := strings.LastIndex(rawURL, "."); dotIndex != -1 {
			rawURL = rawURL[:dotIndex]
		}
		// Decode base64 URL
		decodedBytes, err := base64.URLEncoding.DecodeString(rawURL)
		if err != nil {
			http.Error(w, "Invalid base64 URL", http.StatusBadRequest)
			return
		}
		imageURL = string(decodedBytes)
	} else {
		http.Error(w, "Invalid URL type", http.StatusBadRequest)
		return
	}

	// Pre-process the image URL in case Chrome trimmed one slash
	if strings.HasPrefix(imageURL, "http:/") && !strings.HasPrefix(imageURL, "http://") {
		imageURL = strings.Replace(imageURL, "http:/", "http://", 1)
	} else if strings.HasPrefix(imageURL, "https:/") && !strings.HasPrefix(imageURL, "https://") {
		imageURL = strings.Replace(imageURL, "https:/", "https://", 1)
	} else if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") {
		// If the URL does not have an http or https prefix, add https://
		imageURL = "https://" + imageURL
	}

	// Parse the image URL to handle query parameters
	parsedURL, err := url.Parse(imageURL)
	if err != nil || !parsedURL.IsAbs() {
		http.Error(w, "Invalid image URL", http.StatusBadRequest)
		return
	}

	// Reconstruct the image URL with query parameters
	imageURL = parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + "?" + parsedURL.RawQuery
	if cfg.Debug {
		log.Printf("Fetching image from URL: %s", imageURL)
	}

	// Create a custom HTTP client with insecure SSL verification
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	// Create a new HTTP request with headers to mimic a Chrome browser
	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Fetch the image using the custom client
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error fetching image: %v", err)
		http.Error(w, "Failed to fetch image", http.StatusInternalServerError)
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Printf("Image fetch returned status: %d", resp.StatusCode)
		http.Error(w, "Image not found", http.StatusNotFound)
		return
	}
	defer resp.Body.Close()

	// Decode the image
	img, err := imgconv.Decode(resp.Body)
	if err != nil {
		http.Error(w, "Failed to decode image", http.StatusUnsupportedMediaType)
		return
	}

	// Process image parameters
	width, _ := strconv.Atoi(optionsMap["w"])
	height, _ := strconv.Atoi(optionsMap["h"])
	quality, err := strconv.Atoi(optionsMap["q"])
	if err != nil {
		quality = cfg.Quality
	}
	format := optionsMap["f"]

	// Get the original image dimensions
	originalBounds := img.Bounds()
	originalWidth := originalBounds.Dx()
	originalHeight := originalBounds.Dy()

	// Use original dimensions if specified dimensions are too large
	if width > originalWidth {
		width = originalWidth
	}
	if height > originalHeight {
		height = originalHeight
	}

	// Resize the image if width or height is specified
	if width > 0 || height > 0 {
		// Calculate the aspect ratio of the original image
		originalAspectRatio := float64(originalWidth) / float64(originalHeight)

		// If only one dimension is set, calculate the other to maintain the original aspect ratio
		if width == 0 {
			width = int(float64(height) * originalAspectRatio)
		} else if height == 0 {
			height = int(float64(width) / originalAspectRatio)
		}

		// Calculate the aspect ratio of the target dimensions
		targetAspectRatio := float64(width) / float64(height)

		if originalAspectRatio != targetAspectRatio {
			// Calculate the crop rectangle
			var cropRect image.Rectangle
			if originalAspectRatio > targetAspectRatio {
				// Wider than target, crop width
				newWidth := int(float64(originalHeight) * targetAspectRatio)
				x0 := (originalWidth - newWidth) / 2
				cropRect = image.Rect(x0, 0, x0+newWidth, originalHeight)
			} else {
				// Taller than target, crop height
				newHeight := int(float64(originalWidth) / targetAspectRatio)
				y0 := (originalHeight - newHeight) / 2
				cropRect = image.Rect(0, y0, originalWidth, y0+newHeight)
			}

			// Crop the image
			img = img.(interface {
				SubImage(r image.Rectangle) image.Image
			}).SubImage(cropRect)
		}

		// Resize the image to the target dimensions
		img = imgconv.Resize(img, &imgconv.ResizeOption{Width: width, Height: height})
	}

	// Apply image adjustments using bild library
	// Brightness adjustment (-100 to 100)
	if brightnessStr, ok := optionsMap["b"]; ok {
		brightness, err := strconv.ParseFloat(brightnessStr, 64)
		if err == nil {
			// Convert from -100 to 100 scale to -1 to 1 scale used by bild
			brightness = brightness / 100
			img = adjust.Brightness(img, brightness)
			if cfg.Debug {
				log.Printf("Applied brightness adjustment: %f", brightness)
			}
		}
	}

	// Contrast adjustment (-100 to 100)
	if contrastStr, ok := optionsMap["c"]; ok {
		contrast, err := strconv.ParseFloat(contrastStr, 64)
		if err == nil {
			// Convert from -100 to 100 scale to -1 to 1 scale used by bild
			contrast = contrast / 100
			img = adjust.Contrast(img, contrast)
			if cfg.Debug {
				log.Printf("Applied contrast adjustment: %f", contrast)
			}
		}
	}

	// Gamma adjustment (0 to 10)
	if gammaStr, ok := optionsMap["g"]; ok {
		gamma, err := strconv.ParseFloat(gammaStr, 64)
		if err == nil {
			img = adjust.Gamma(img, gamma)
			if cfg.Debug {
				log.Printf("Applied gamma adjustment: %f", gamma)
			}
		}
	}

	// Hue adjustment (-360 to 360 degrees)
	if hueStr, ok := optionsMap["h"]; ok {
		hue, err := strconv.Atoi(hueStr)
		if err == nil {
			img = adjust.Hue(img, hue)
			if cfg.Debug {
				log.Printf("Applied hue adjustment: %d", hue)
			}
		}
	}

	// Saturation adjustment (-100 to 100)
	if saturationStr, ok := optionsMap["s"]; ok {
		saturation, err := strconv.ParseFloat(saturationStr, 64)
		if err == nil {
			// Convert from -100 to 100 scale to -1 to 1 scale used by bild
			saturation = saturation / 100
			img = adjust.Saturation(img, saturation)
			if cfg.Debug {
				log.Printf("Applied saturation adjustment: %f", saturation)
			}
		}
	}

	// Black and white conversion
	if bwStr, ok := optionsMap["bw"]; ok {
		if bwStr == "1" {
			img = effect.Grayscale(img)
			if cfg.Debug {
				log.Printf("Applied black and white conversion")
			}
		}
	}

	// Sharpen effect
	if shrpStr, ok := optionsMap["shrp"]; ok {
		if shrpStr == "1" {
			img = effect.Sharpen(img)
			if cfg.Debug {
				log.Printf("Applied sharpen effect")
			}
		}
	}

	// Check if the 'nw' query parameter is present to disable watermark
	_, present := optionsMap["nw"]
	disableWatermark := present

	// Apply watermark if watermark image path is set and 'nw' is not present
	if watermarkConfig.Image != "" && (!disableWatermark || watermarkConfig.Force) {
		// Load the watermark image
		watermarkImg, err := imgconv.Open(watermarkConfig.Image)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load watermark image: %s", watermarkConfig.Image),
				http.StatusInternalServerError)
			return
		}

		// Resize the watermark based on the percentage of the original image
		imgBounds := img.Bounds()
		maxWatermarkWidth := imgBounds.Dx() * watermarkConfig.SizePercent / 100
		maxWatermarkHeight := imgBounds.Dy() * watermarkConfig.SizePercent / 100

		// Calculate the aspect ratio of the watermark
		watermarkBounds := watermarkImg.Bounds()
		watermarkAspectRatio := float64(watermarkBounds.Dx()) / float64(watermarkBounds.Dy())

		// Determine the new dimensions while maintaining aspect ratio
		var newWatermarkWidth, newWatermarkHeight int
		if float64(maxWatermarkWidth)/watermarkAspectRatio <= float64(maxWatermarkHeight) {
			newWatermarkWidth = maxWatermarkWidth
			newWatermarkHeight = int(float64(maxWatermarkWidth) / watermarkAspectRatio)
		} else {
			newWatermarkHeight = maxWatermarkHeight
			newWatermarkWidth = int(float64(maxWatermarkHeight) * watermarkAspectRatio)
		}

		watermarkImg = imgconv.Resize(watermarkImg, &imgconv.ResizeOption{Width: newWatermarkWidth, Height: newWatermarkHeight})

		watermarkOption := &imgconv.WatermarkOption{
			Mark:    watermarkImg,
			Opacity: uint8(watermarkConfig.Opacity * 255 / 100), // Convert percentage to 0-255 scale
		}

		if watermarkConfig.Random {
			watermarkOption.SetRandom(true)
		} else {
			// Get position from query parameter or use config default
			position := optionsMap["position"]
			if position == "" {
				position = watermarkConfig.Position
			}

			// Calculate offsets based on position and percentage
			watermarkBounds := watermarkImg.Bounds()
			offsetX, offsetY := calculateWatermarkOffset(
				position,
				imgBounds.Dx(),
				imgBounds.Dy(),
				watermarkBounds.Dx(),
				watermarkBounds.Dy(),
				watermarkConfig.OffsetXPercent,
				watermarkConfig.OffsetYPercent,
			)

			watermarkOption.SetOffset(image.Pt(offsetX, offsetY))
		}

		img = imgconv.Watermark(img, watermarkOption)
	}

	// Set the output format
	var encodeFunc func(io.Writer, image.Image) error
	switch strings.ToLower(format) {
	case "png":
		encodeFunc = func(w io.Writer, img image.Image) error {
			return png.Encode(w, img)
		}
	default:
		// default to jpeg
		encodeFunc = func(w io.Writer, img image.Image) error {
			return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})
		}
	}

	// Encode and write the image
	err = encodeFunc(w, img)
	if err != nil {
		http.Error(w, "Failed to encode image", http.StatusInternalServerError)
		return
	}
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	if cfg.Debug {
		log.Printf("Config: %v", cfg)
	}

	// Load watermark configurations
	watermarkConfigs, err := loadWatermarkConfigs(cfg.WatermarksConfigFile)
	if err != nil {
		log.Fatalf("Failed to load watermark configurations: %v", err)
	}

	if cfg.Debug {
		log.Printf("Watermark configurations: %v", watermarkConfigs)
	}

	// Pass watermarkConfigs to the handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		processImage(w, r, cfg, watermarkConfigs)
	})

	log.Printf("Starting server on :%d", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
}
