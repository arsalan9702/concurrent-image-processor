package processor

import (
	"context"
	"sync"
	"time"

	"github.com/arsalan9702/concurrent-image-processor/internal/models"
	"github.com/arsalan9702/concurrent-image-processor/pkg/logger"
)

// manage pool of workers for jobs
type WorkerPool struct {
	workerCount int
	jobQueue    chan models.ImageJob
	resultQueue chan models.ProcessingResult
	rowResults  chan models.RowResult
	quit        chan bool
	wg          sync.WaitGroup
	logger      logger.Logger
	processor   *Processor
}

// create new worker pool
func NewWorkerPool(workerCount int, bufferSize int, log logger.Logger, processor *Processor) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		jobQueue:    make(chan models.ImageJob, bufferSize),
		resultQueue: make(chan models.ProcessingResult, bufferSize),
		rowResults:  make(chan models.RowResult, bufferSize*10),
		quit:        make(chan bool),
		logger:      log,
		processor:   processor,
	}
}

// intitalize and start workers
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.logger.WithField("workers", wp.workerCount).Info("Starting worker pool")

	for i := 0; i < wp.workerCount; i++ {
		wp.wg.Add(1)
		go wp.imageWorker(ctx, i)
	}
}

// gracefully stop workers
func (wp *WorkerPool) Stop() {
	wp.logger.Info("Stopping worker pool")
	close(wp.quit)
	close(wp.jobQueue)
	wp.wg.Wait()
	close(wp.resultQueue)
}

// submit an image processing job
func (wp *WorkerPool) SubmitJob(job models.ImageJob) {
	select {
	case wp.jobQueue <- job:
	case <-wp.quit:
		wp.logger.Warn("Worker pool shutting down, job rejected")
	}
}

// return the results channel
func (wp *WorkerPool) Results() <-chan models.ProcessingResult {
	return wp.resultQueue
}

// process image jobs
func (wp *WorkerPool) imageWorker(ctx context.Context, workerID int) {
	defer wp.wg.Done()

	log := wp.logger.WithField("worker_id", workerID)
	log.Debug("Image worker started")

	for {
		select {
		case <-ctx.Done():
			log.Debug("Image worker stopped due to context cancellation")
			return
		case <-wp.quit:
			log.Debug("Image worker stopped")
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				log.Debug("Image worker stopped, job queue closed")
				return
			}

			log.WithFields(map[string]interface{}{
				"job_id":     job.ID,
				"input_path": job.InputPath,
				"filter":     job.Filter,
			}).Debug("Processing image job")

			result := wp.processor.ProcessSingleImage(ctx, job)

			select {
			case wp.resultQueue <- result:
			case <-ctx.Done():
				return
			case <-wp.quit:
				return
			}
		}
	}
}

// ImageProcessor handles the actual image processing logic
type ImageProcessor struct {
	logger logger.Logger
}

// ProcessSingleImage processes a single image job (placeholder implementation)
func (ip *ImageProcessor) ProcessSingleImage(ctx context.Context, job models.ImageJob) models.ProcessingResult {
	startTime := time.Now()

	result := models.ProcessingResult{
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
	}

	// In the actual implementation, this would delegate to the main processor
	// For now, simulate processing
	select {
	case <-ctx.Done():
		result.Error = ctx.Err()
		return result
	case <-time.After(time.Millisecond * 100): // Simulate work
	}

	result.ProcessingTime = time.Since(startTime)
	result.Metadata = models.ImageMetadata{
		Width:         800, // Simulated values
		Height:        600,
		Format:        "jpeg",
		OriginalSize:  102400,
		ProcessedSize: 98304,
		RowsProcessed: 600,
	}

	return result
}
