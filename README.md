# Concurrent Image Processor

A high-performance, concurrent image processing application written in Go that applies various filters to images using goroutines for parallel processing.

## Features

- **Concurrent Processing**: Utilizes worker pools and row-level parallelism
- **Multiple Filters**: Supports grayscale, blur, brightness, and contrast filters
- **Multiple Formats**: Handles JPEG, PNG, GIF, BMP, TIFF, and WebP images
- **Configurable**: Supports configuration files and command-line arguments
- **Logging**: Comprehensive logging with configurable verbosity
- **Graceful Shutdown**: Handles interruption signals properly

## Installation

```bash
# Clone the repository
git clone https://github.com/arsalan9702/concurrent-image-processor.git
cd concurrent-image-processor

# Initialize Go module
go mod init github.com/arsalan9702/concurrent-image-processor

# Install dependencies
go mod tidy

# Build the application
go build -o bin/processor cmd/processor/main.go
```

## Usage

### Basic Usage

```bash
# Process images with default settings (grayscale filter)
./bin/processor -input examples/images -output examples/output

# Apply different filters
./bin/processor -input examples/images -output examples/output -filter blur
./bin/processor -input examples/images -output examples/output -filter brightness
./bin/processor -input examples/images -output examples/output -filter contrast

# Specify number of workers
./bin/processor -input examples/images -output examples/output -workers 8

# Enable verbose logging
./bin/processor -input examples/images -output examples/output -verbose
```

### Command Line Options

- `-input`: Input directory containing images (default: "examples/images")
- `-output`: Output directory for processed images (default: "examples/output")
- `-filter`: Filter to apply - grayscale, blur, brightness, contrast (default: "grayscale")
- `-workers`: Number of worker goroutines (default: number of CPU cores)
- `-row-workers`: Number of row processing workers per image (default: CPU cores * 2)
- `-config`: Configuration file path
- `-verbose`: Enable verbose logging

### Configuration File

Create a YAML configuration file:

```yaml
input_dir: "examples/images"
output_dir: "examples/output"
filter: "grayscale"
workers: 4
row_workers: 8
quality: 95
blur_radius: 2.0
brightness: 1.2
contrast: 1.1
max_file_size: 104857600  # 100MB
buffer_size: 1000
```

Use with: `./bin/processor -config config.yaml`

### Environment Variables

Set environment variables with `IMG_PROC_` prefix:

```bash
export IMG_PROC_INPUT_DIR="examples/images"
export IMG_PROC_OUTPUT_DIR="examples/output"
export IMG_PROC_FILTER="blur"
export IMG_PROC_WORKERS=8
```

## Architecture

### Components

1. **Main**: Entry point, handles CLI arguments and orchestrates processing
2. **Config**: Configuration management with defaults, file, and environment support
3. **Processor**: Core image processing logic with worker pool management
4. **Worker Pool**: Concurrent job processing using goroutines
5. **Filters**: Image filter implementations (grayscale, blur, brightness, contrast)
6. **Models**: Data structures for jobs, results, and metadata
7. **Logger**: Structured logging with configurable levels

### Processing Flow

1. **Discovery**: Find all supported image files in input directory
2. **Job Creation**: Create processing jobs for each image
3. **Worker Pool**: Distribute jobs across worker goroutines
4. **Row Processing**: Each image is processed row by row in parallel
5. **Filter Application**: Apply selected filter to pixel data
6. **Output**: Save processed images to output directory

## Supported Image Formats

- JPEG (.jpg, .jpeg)
- PNG (.png)
- GIF (.gif)
- BMP (.bmp)
- TIFF (.tiff, .tif)
- WebP (.webp)

## Available Filters

### Grayscale
Converts images to grayscale using standard luminance formula: `0.299*R + 0.587*G + 0.114*B`

### Blur
Applies box blur filter with configurable radius.

### Brightness
Adjusts image brightness by multiplying RGB values by a factor.

### Contrast
Adjusts image contrast by scaling RGB values around midpoint (128).

## Performance

The application is designed for high performance:

- **Concurrent Processing**: Multiple images processed simultaneously
- **Row-Level Parallelism**: Each image row processed in parallel
- **Efficient Memory Usage**: Processes images in chunks
- **Configurable Workers**: Tune for your hardware

## Building and Development

### Project Structure

```
concurrent-image-processor/
├── cmd/processor/          # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── models/            # Data structures
│   └── processor/         # Core processing logic
├── pkg/logger/            # Logging utilities
├── scripts/               # Build and test scripts
├── examples/              # Example images and outputs
└── README.md
```

### Build Scripts

```bash
# Make scripts executable
chmod +x scripts/build.sh scripts/test.sh

# Build the application
./scripts/build.sh

# Run tests
./scripts/test.sh
```

## Examples

Create an `examples/images` directory and add some test images:

```bash
mkdir -p examples/images examples/output

# Add your images to examples/images/
# Then run:
./bin/processor
```
