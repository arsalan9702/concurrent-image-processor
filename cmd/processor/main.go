package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/arsalan9702/concurrent-image-processor/internal/config"
	"github.com/arsalan9702/concurrent-image-processor/internal/processor"
	"github.com/arsalan9702/concurrent-image-processor/pkg/logger"
)

func main() {
	var (
		inputDir   = flag.String("input", "examples/images", "Input directory containing images")
		outputDir  = flag.String("output", "examples/output", "Output directory for processed images")
		filter     = flag.String("filter", "grayscale", "Filter to apply (grayscale, blur, birghtness, contrast)")
		workers    = flag.Int("workers", runtime.NumCPU(), "Number of worker goroutines")
		rowWorkers = flag.Int("row-workers", runtime.NumCPU()*2, "Number of row processing workers per image")
		configFile = flag.String("config", "", "Configuration file path")
		verbose    = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	log:=logger.NewLogger(*verbose)

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.WithError(err).Fatal("Failed to load config file")
	}

	if *inputDir!="examples/images"{
		cfg.InputDir = *inputDir
	}
	if *outputDir!="examples/output"{
		cfg.OutputDir = *outputDir
	}
	if *filter!="grayscale"{
		cfg.Filter = *filter
	}
	if *workers!=runtime.NumCPU(){
		cfg.Workers = *workers
	}
	if *rowWorkers!=runtime.NumCPU()*2{
		cfg.RowWorkers = *rowWorkers
	}

	log.WithFields(map[string]interface{}{
		"input_dir":   cfg.InputDir,
		"output_dir":  cfg.OutputDir,
		"filter":      cfg.Filter,
		"workers":     cfg.Workers,
		"row_workers": cfg.RowWorkers,
	}).Info("Starting image processor")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func(){
		<-sigChan
		log.Info("Received shutdown signal, stopping")
		cancel()
	}()

	if err:=os.MkdirAll(cfg.OutputDir, 0755);err!=nil{
		log.WithError(err).Fatal("Failed to create output directory")
	}

	proc, err:= processor.New(cfg, log)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize processor")
	}

	imageFiles, err:= findImageFiles(cfg.InputDir)
	if err != nil {
		log.WithError(err).Fatal("No images found in input directory")
	}

	if len(imageFiles)==0{
		log.Warn("No images found in input directory")
		return
	}

	log.WithField("count", len(imageFiles)).Info("Found image files")

	startTime:=time.Now()
	results, err:= proc.ProcessImages(ctx, imageFiles)
	if err != nil {
		log.WithError(err).Fatal("Failed to process images")
	}

	duration:=time.Since(startTime)
	successful:=0
	failed:=0

	for _, result := range results {
		if result.Error != nil {
			log.WithError(result.Error).WithField("file", result.InputPath).Error("failed to process image")
			failed++
		} else {
			log.WithFields(map[string]interface{}{
				"input": result.InputPath,
				"output": result.OutputPath,
				"duration": result.ProcessingTime,
			}).Info("Successfully processed image")
			successful++
		}
	}

	log.WithFields(map[string]interface{}{
		"total_duration": duration,
		"successful":     successful,
		"failed":         failed,
		"total":          len(results),
	}).Info("Processing completed")
}

func findImageFiles(dir string) ([]string, error) {
	var files []string
	supportedExts:=map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".tiff": true,
		".webp": true,
	}

	err:=filepath.Walk(dir, func(path string, info os.FileInfo, err error) error{
		if err != nil {
			return nil
		}

		if !info.IsDir() {
			ext:=strings.ToLower(filepath.Ext(path))
			if supportedExts[ext]{
				files=append(files, path)
			}
		}

		return nil
	})

	return files, err
}

