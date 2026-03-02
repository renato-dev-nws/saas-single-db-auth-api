package main

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"github.com/saas-single-db-api/internal/config"
	"github.com/saas-single-db-api/internal/database"
	"github.com/saas-single-db-api/internal/storage"
)

// Variant config
type variantConfig struct {
	Name      string
	MaxWidth  int
	MaxHeight int
}

var variants = []variantConfig{
	{Name: "medium", MaxWidth: 800, MaxHeight: 800},
	{Name: "small", MaxWidth: 350, MaxHeight: 350},
	{Name: "thumb", MaxWidth: 100, MaxHeight: 100},
}

// imageRow holds the fields we read from DB
type imageRow struct {
	ID               string
	TenantID         string
	ImageableType    string
	ImageableID      string
	OriginalFilename *string
	MimeType         string
	Extension        string
	StorageDriver    string
	OriginalPath     string
	OriginalURL      *string
	ProcessingStatus string
	FileSize         *int64
}

type worker struct {
	db      *pgxpool.Pool
	storage storage.Provider
	rdb     *redis.Client
}

func main() {
	log.Println("🖼️  Image Worker starting...")

	cfg := config.Load()

	db := database.NewPostgresPool(cfg.DatabaseURL)
	defer db.Close()

	storageProvider, err := storage.NewProvider(cfg)
	if err != nil {
		log.Fatalf("Failed to create storage provider: %v", err)
	}

	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(opts)
	defer rdb.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✓ Connected to Redis")

	w := &worker{db: db, storage: storageProvider, rdb: rdb}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()

	go func() {
		<-sigCh
		log.Println("Shutting down worker...")
		bgCancel()
	}()

	w.subscribe(bgCtx)
}

func (w *worker) subscribe(ctx context.Context) {
	pubsub := w.rdb.Subscribe(ctx, "image:process")
	defer pubsub.Close()

	log.Println("✓ Subscribed to channel: image:process")

	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker context cancelled, exiting")
			return
		case msg, ok := <-ch:
			if !ok {
				log.Println("Channel closed, exiting")
				return
			}
			w.handleMessage(ctx, msg.Payload)
		}
	}
}

func (w *worker) handleMessage(ctx context.Context, payload string) {
	var event struct {
		ImageID string `json:"image_id"`
	}
	if err := json.Unmarshal([]byte(payload), &event); err != nil {
		log.Printf("Error parsing message: %v", err)
		return
	}

	log.Printf("Processing image: %s", event.ImageID)

	if err := w.processImage(ctx, event.ImageID); err != nil {
		log.Printf("Error processing image %s: %v", event.ImageID, err)
	} else {
		log.Printf("Image %s processed successfully", event.ImageID)
	}
}

func (w *worker) processImage(ctx context.Context, imageID string) error {
	// 1. Fetch image record
	img, err := w.getImage(ctx, imageID)
	if err != nil {
		return fmt.Errorf("failed to get image: %w", err)
	}

	// 2. Validate
	if img.ProcessingStatus != "pending" {
		return fmt.Errorf("image not eligible: status=%s", img.ProcessingStatus)
	}

	// 3. Set status to processing
	if err := w.updateStatus(ctx, imageID, "processing"); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// 4. Check tenant convert_webp setting (default true)
	convertWebp := w.getConvertWebp(ctx, img.TenantID)

	// 5. Download original from storage
	reader, err := w.storage.GetReader(img.OriginalPath)
	if err != nil {
		w.updateStatus(ctx, imageID, "failed")
		return fmt.Errorf("failed to get reader: %w", err)
	}
	defer reader.Close()

	// 6. Decode image
	srcImage, format, err := image.Decode(reader)
	if err != nil {
		w.updateStatus(ctx, imageID, "failed")
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// 7. Update original dimensions
	bounds := srcImage.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()
	w.updateDimensions(ctx, imageID, origWidth, origHeight)

	// 8. Generate variants and update the same row
	for _, v := range variants {
		variantPath, variantURL, err := w.generateVariant(ctx, img, srcImage, format, v, convertWebp)
		if err != nil {
			log.Printf("Error generating %s variant for image %s: %v", v.Name, imageID, err)
			w.updateStatus(ctx, imageID, "failed")
			return fmt.Errorf("failed to generate variant %s: %w", v.Name, err)
		}
		// Update the image row with the variant path and URL
		w.updateVariant(ctx, imageID, v.Name, variantPath, variantURL)
	}

	// 9. Mark as completed
	if err := w.updateStatusCompleted(ctx, imageID); err != nil {
		return fmt.Errorf("failed to mark completed: %w", err)
	}

	return nil
}

func (w *worker) generateVariant(ctx context.Context, img *imageRow, srcImage image.Image, format string, vc variantConfig, convertWebp bool) (string, string, error) {
	// Resize using Lanczos
	resized := imaging.Fit(srcImage, vc.MaxWidth, vc.MaxHeight, imaging.Lanczos)

	// Determine output format and extension
	outExt := img.Extension
	if convertWebp {
		outExt = "webp"
	}

	// Build variant filename from original path
	origBase := filepath.Base(img.OriginalPath)
	nameWithoutExt := strings.TrimSuffix(origBase, filepath.Ext(origBase))
	variantFilename := fmt.Sprintf("%s_%s.%s", nameWithoutExt, vc.Name, outExt)

	// Build storage path from original path
	dir := filepath.Dir(img.OriginalPath)
	variantStoragePath := filepath.Join(dir, variantFilename)

	// Get full filesystem path for local storage
	storagePath := cfg_storagePath()
	fullPath := filepath.Join(storagePath, variantStoragePath)

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", "", fmt.Errorf("failed to create dir: %w", err)
	}

	// Create file
	outFile, err := os.Create(fullPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer outFile.Close()

	// Encode based on output format
	if convertWebp {
		err = webp.Encode(outFile, resized, &webp.Options{Lossless: false, Quality: 85})
	} else {
		switch format {
		case "jpeg", "jpg":
			err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 90})
		case "png":
			err = png.Encode(outFile, resized)
		default:
			err = jpeg.Encode(outFile, resized, &jpeg.Options{Quality: 90})
		}
	}
	if err != nil {
		os.Remove(fullPath)
		return "", "", fmt.Errorf("failed to encode: %w", err)
	}

	// Build public URL
	baseURL := cfg_storageBaseURL()
	variantURL := fmt.Sprintf("%s/%s", baseURL, variantStoragePath)

	return variantStoragePath, variantURL, nil
}

// updateVariant updates the image row with the path and URL for a specific variant (medium, small, thumb).
func (w *worker) updateVariant(ctx context.Context, imageID, variantName, variantPath, variantURL string) {
	var col string
	switch variantName {
	case "medium":
		col = "medium"
	case "small":
		col = "small"
	case "thumb":
		col = "thumb"
	default:
		return
	}
	query := fmt.Sprintf(`UPDATE images SET %s_path = $1, %s_url = $2, updated_at = NOW() WHERE id = $3`, col, col)
	w.db.Exec(ctx, query, variantPath, variantURL, imageID)
}

func (w *worker) getImage(ctx context.Context, imageID string) (*imageRow, error) {
	var img imageRow
	err := w.db.QueryRow(ctx,
		`SELECT id, tenant_id, imageable_type, imageable_id, original_filename, mime_type, extension, storage_driver, original_path, original_url, processing_status, file_size
		 FROM images WHERE id = $1`, imageID,
	).Scan(&img.ID, &img.TenantID, &img.ImageableType, &img.ImageableID, &img.OriginalFilename, &img.MimeType, &img.Extension, &img.StorageDriver, &img.OriginalPath, &img.OriginalURL, &img.ProcessingStatus, &img.FileSize)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (w *worker) updateStatus(ctx context.Context, imageID, status string) error {
	_, err := w.db.Exec(ctx,
		`UPDATE images SET processing_status = $1, updated_at = NOW() WHERE id = $2`, status, imageID,
	)
	return err
}

func (w *worker) updateStatusCompleted(ctx context.Context, imageID string) error {
	_, err := w.db.Exec(ctx,
		`UPDATE images SET processing_status = 'completed', processed_at = NOW(), updated_at = NOW() WHERE id = $1`, imageID,
	)
	return err
}

func (w *worker) updateDimensions(ctx context.Context, imageID string, width, height int) {
	w.db.Exec(ctx,
		`UPDATE images SET width = $1, height = $2, updated_at = NOW() WHERE id = $3`, width, height, imageID,
	)
}

func (w *worker) getConvertWebp(ctx context.Context, tenantID string) bool {
	var convertWebp bool
	err := w.db.QueryRow(ctx,
		`SELECT convert_webp FROM tenant_settings WHERE tenant_id = $1`, tenantID,
	).Scan(&convertWebp)
	if err != nil {
		// Default to true if no settings row exists
		return true
	}
	return convertWebp
}

// Config helpers — read from environment with defaults
func cfg_storagePath() string {
	if v := os.Getenv("STORAGE_LOCAL_PATH"); v != "" {
		return v
	}
	return "./uploads"
}

func cfg_storageBaseURL() string {
	if v := os.Getenv("STORAGE_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080/uploads"
}
