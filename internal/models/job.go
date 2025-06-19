package models

import (
	"image"
	"time"
)

type FilterType string

const (
	FilterGrayScale  FilterType = "grayscale"
	FilterBlur       FilterType = "blur"
	FilterBrightness FilterType = "brightness"
	FilterConstrast  FilterType = "contrast"
)

// single image processing job
type ImageJob struct {
	ID         string
	InputPath  string
	OutputPath string
	Filter     FilterType
	Params     FilterParams
}

// parameters for different filters
type FilterParams struct {
	BlurRadius float64
	Brightness float64
	Contrast   float64
	Quality    int
}

// result of processing image
type ProcessingResult struct {
	InputPath      string
	OutputPath     string
	ProcessingTime time.Duration
	Error          error
	Metadata       ImageMetadata
}

// info of processed image
type ImageMetadata struct {
	Width         int
	Height        int
	Format        string
	OriginalSize  int64
	ProcessedSize int64
	RowsProcessed int
}

// job for processing a single row
type RowJob struct{
	ImageID string
	RowIndex int
	Pixels []uint8
	Width int
	Bounds image.Rectangle
	Filter FilterType
	Params FilterParams
}

// result of processing single row
type RowResult struct{
	ImageID string
	RowIndex int
	Pixels []uint8
	Error error
	Duration time.Duration
}

