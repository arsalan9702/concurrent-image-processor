package processor

import (
	"context"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"

	
	"image/jpeg"
	"image/png"

	"github.com/arsalan9702/concurrent-image-processor/internal/config"
	"github.com/arsalan9702/concurrent-image-processor/internal/models"
	"github.com/arsalan9702/concurrent-image-processor/pkg/logger"
)

// handles current image processing
type Processor struct {
	config     *config.Config
	workerPool *WorkerPool
	logger     logger.Logger
}

// create new processor instance
func New(cfg *config.Config, log logger.Logger) (*Processor, error) {
	processor := &Processor{
		config: cfg,
		logger: log,
	}
	
	// Pass the processor instance to the worker pool
	workerPool := NewWorkerPool(cfg.Workers, cfg.BufferSize, log, processor)
	processor.workerPool = workerPool

	return processor, nil
}

// process multiple images concurrently
func (p *Processor) ProcessImages(ctx context.Context, imagePaths []string) ([]models.ProcessingResult, error) {
	p.logger.WithField("count", len(imagePaths)).Info("Starting batch image processing")

	p.workerPool.Start(ctx)
	defer p.workerPool.Stop()

	for i, path := range imagePaths {
		job := models.ImageJob{
			ID:         fmt.Sprintf("job_%d", i),
			InputPath:  path,
			OutputPath: p.generateOutputPath(path),
			Filter:     models.FilterType(p.config.Filter),
			Params: models.FilterParams{
				BlurRadius: p.config.BlurRadius,
				Brightness: p.config.Brightness,
				Contrast:   p.config.Contrast,
				Quality:    p.config.Quality,
			},
		}

		p.workerPool.SubmitJob(job)
	}

	var results []models.ProcessingResult
	resultsReceived := 0
	expectedResults := len(imagePaths)

	for resultsReceived < expectedResults {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		case result := <-p.workerPool.Results():
			results = append(results, result)
			resultsReceived++
		}
	}

	return results, nil
}

// process single image with row-level concurrency
func (p *Processor) ProcessSingleImage(ctx context.Context, job models.ImageJob) models.ProcessingResult {
	startTime := time.Now()
	log := p.logger.WithFields(map[string]interface{}{
		"job_id":     job.ID,
		"input_path": job.InputPath,
		"filter":     job.Filter,
	})

	result := models.ProcessingResult{
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
	}

	// check file size
	fileInfo, err := os.Stat(job.InputPath)
	if err != nil {
		result.Error = fmt.Errorf("fialed to stat file: %w", err)
		return result
	}

	if fileInfo.Size() > p.config.MaxFileSize {
		result.Error = fmt.Errorf("file size %d exceeds maximum %d", fileInfo.Size(), p.config.MaxFileSize)
		return result
	}

	result.Metadata.OriginalSize = fileInfo.Size()

	img, format, err := p.loadImage(job.InputPath)
	if err != nil {
		result.Error = fmt.Errorf("failed to load image: %w", err)
		return result
	}

	log.WithFields(map[string]interface{}{
		"width":  img.Bounds().Dx(),
		"height": img.Bounds().Dy(),
		"format": format,
	}).Debug("Image loaded successfully")

	rgba := ImageToRGBA(img)
	bounds := rgba.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	result.Metadata.Width = width
	result.Metadata.Height = height
	result.Metadata.Format = format
	result.Metadata.RowsProcessed = height

	// process image row by row using goroutines
	processedRows := make([][]uint8, height)
	var wg sync.WaitGroup
	rowResults := make(chan models.RowResult, height)

	for row := 0; row < height; row++ {
		wg.Add(1)
		go func(rowIndex int) {
			defer wg.Done()

			pixels := ExtractRowPixels(rgba, rowIndex)
			if pixels == nil {
				rowResults <- models.RowResult{
					ImageID:  job.ID,
					RowIndex: rowIndex,
					Error:    fmt.Errorf("failed to extract pixels for row %d", rowIndex),
				}
				return
			}

			var processPixels []uint8
			if filter, exists := FilterRegistry[job.Filter]; exists {
				processPixels = filter(pixels, width, job.Params)
			} else {
				rowResults <- models.RowResult{
					ImageID:  job.ID,
					RowIndex: rowIndex,
					Error:    fmt.Errorf("unknown filter: %s", job.Filter),
				}
				return
			}

			rowResults <- models.RowResult{
				ImageID:  job.ID,
				RowIndex: rowIndex,
				Pixels:   processPixels,
				Error:    nil,
			}
		}(row)
	}

	go func() {
		wg.Wait()
		close(rowResults)
	}()

	// collect row results
	for rowResult := range rowResults {
		if rowResult.Error != nil {
			result.Error = fmt.Errorf("row processing failed: %w", rowResult.Error)
			return result
		}
		processedRows[rowResult.RowIndex] = rowResult.Pixels
	}

	for row := 0; row < height; row++ {
		if processedRows[row] != nil {
			SetRowPixels(rgba, row, processedRows[row])
		}
	}

	if err := p.saveImage(rgba, job.OutputPath, format, job.Params.Quality); err != nil {
		result.Error = fmt.Errorf("failed to save image: %w", err)
		return result
	}

	if outputInfo, err := os.Stat(job.OutputPath); err != nil {
		result.Metadata.ProcessedSize = outputInfo.Size()
	}

	result.ProcessingTime = time.Since(startTime)
	log.WithField("duration", result.ProcessingTime).Info("image processing completed")

	return result
}

// loading image
func (p *Processor) loadImage(path string) (image.Image, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}

	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".webp":
		img, err := webp.Decode(file)
		return img, "webp", err
	case ".bmp":
		img, err := bmp.Decode(file)
		return img, "bmp", err
	case ".tiff", ".tif":
		img, err := tiff.Decode(file)
		return img, "tiff", err
	default:
		// Use Go's built-in image decoder
		img, format, err := image.Decode(file)
		return img, format, err
	}
}

func (p *Processor) saveImage(img image.Image, path string, originalFormat string, quality int) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	format := originalFormat

	if ext == ".jpg" || ext == ".jpeg" {
		format = "jpeg"
	} else if ext == ".png" {
		format = "png"
	}

	switch format{
		case "jpeg":
			options := &jpeg.Options{Quality: quality}
			return jpeg.Encode(file, img, options)
		case "png":
			encoder:= &png.Encoder{CompressionLevel: png.BestCompression}
			return encoder.Encode(file, img)
		default:
			encoder:= &png.Encoder{CompressionLevel: png.BestCompression}
			return encoder.Encode(file, img)
	}
}

func (p *Processor) generateOutputPath(inputPath string) string{
	dir := filepath.Dir(inputPath)
	filename:=filepath.Base(inputPath)
	ext:=filepath.Ext(inputPath)
	name:=strings.TrimSuffix(filename, ext)

	outputDir := p.config.OutputDir
	if outputDir == "" {
		outputDir = dir
	}

	outputFilename:= fmt.Sprintf("%s_%s%s", name, p.config.Filter, ext)
	return filepath.Join(outputDir, outputFilename)
}
