package handlers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/meowrain/localsend-go/internal/config"
	"github.com/meowrain/localsend-go/internal/discovery"
	"github.com/meowrain/localsend-go/internal/discovery/shared"
	"github.com/meowrain/localsend-go/internal/models"
	"github.com/meowrain/localsend-go/internal/tui"
	"github.com/meowrain/localsend-go/internal/utils/logger"
	"github.com/meowrain/localsend-go/internal/utils/sha256"
	"github.com/schollz/progressbar/v3"
)

// SendFileToOtherDevicePrepare prepares file metadata and sends it to the target device
func SendFileToOtherDevicePrepare(ip string, path string) (*models.PrepareReceiveResponse, error) {
	// Prepare metadata for all files
	files := make(map[string]models.FileInfo)
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			sha256Hash, err := sha256.CalculateSHA256(filePath)
			if err != nil {
				return fmt.Errorf("error calculating SHA256 hash: %w", err)
			}
			fileMetadata := models.FileInfo{
				ID:       info.Name(), // Use filename as ID
				FileName: info.Name(),
				Size:     info.Size(),
				FileType: filepath.Ext(filePath),
				SHA256:   sha256Hash,
			}
			files[fileMetadata.ID] = fileMetadata
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking the path: %w", err)
	}

	// Create and populate the PrepareReceiveRequest struct
	request := models.PrepareReceiveRequest{
		Info: models.Info{
			Alias:       shared.Message.Alias,
			Version:     shared.Message.Version,
			DeviceModel: shared.Message.DeviceModel,
			DeviceType:  shared.Message.DeviceType,
			Fingerprint: shared.Message.Fingerprint,
			Port:        shared.Message.Port,
			Protocol:    shared.Message.Protocol,
			Download:    shared.Message.Download,
		},
		Files: files,
	}

	// Encode the request struct to JSON
	requestJson, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error encoding request to JSON: %w", err)
	}

	// Send POST request
	url := fmt.Sprintf("https://%s:53317/api/localsend/v2/prepare-upload", ip)
	client := &http.Client{
		Timeout: 60 * time.Second, // Transfer timeout
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip TLS verification for local network
			},
		},
	}
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, fmt.Errorf("error sending POST request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 204:
			return nil, fmt.Errorf("finished (No file transfer needed)")
		case 400:
			return nil, fmt.Errorf("invalid body")
		case 403:
			return nil, fmt.Errorf("rejected")
		case 500:
			return nil, fmt.Errorf("unknown error by receiver")
		}
		return nil, fmt.Errorf("failed to send metadata: received status code %d", resp.StatusCode)
	}

	// Decode response JSON into PrepareReceiveResponse struct
	var prepareReceiveResponse models.PrepareReceiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&prepareReceiveResponse); err != nil {
		return nil, fmt.Errorf("error decoding response JSON: %w", err)
	}

	return &prepareReceiveResponse, nil
}

// uploadFile uploads a single file to the target device
func uploadFile(ctx context.Context, ip, sessionId, fileId, token, filePath string) error {
	// Open the file to send
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Get file size for progress bar
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Create progress bar
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription(fmt.Sprintf("Uploading %s", filepath.Base(filePath))),
		progressbar.OptionSetWidth(15),
		progressbar.OptionShowBytes(true),
		progressbar.OptionThrottle(time.Second), // Reduce refresh rate to minimize flicker
		progressbar.OptionShowCount(),
		progressbar.OptionClearOnFinish(), // Clear progress bar on completion
		progressbar.OptionSetRenderBlankState(true),
		progressbar.OptionSetPredictTime(true), // Predict remaining time
		progressbar.OptionFullWidth(),          // Use full width display
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "█",
			SaucerPadding: "░",
			BarStart:      "|",
			BarEnd:        "|",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	// Build the file upload URL
	uploadURL := fmt.Sprintf("https://%s:53317/api/localsend/v2/upload?sessionId=%s&fileId=%s&token=%s",
		ip, sessionId, fileId, token)

	// Use a pipe to avoid loading the entire file into memory
	pr, pw := io.Pipe()

	// Create an error channel to propagate upload errors
	uploadErr := make(chan error, 1)

	go func() {
		defer pw.Close()
		// Write file data in a goroutine
		_, err := io.Copy(io.MultiWriter(pw, bar), file)
		if err != nil {
			uploadErr <- err
			return
		}
	}()

	// Create HTTP client with TLS config
	client := &http.Client{
		Timeout: 30 * time.Minute,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Skip certificate verification for local network
			},
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
		},
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, pr)
	if err != nil {
		return fmt.Errorf("error creating POST request: %w", err)
	}

	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = fileSize

	// Send request using custom client instead of http.DefaultClient
	resp, err := client.Do(req)

	// Check if cancelled
	select {
	case <-ctx.Done():
		return fmt.Errorf("transfer cancelled")
	case err := <-uploadErr:
		if err != nil {
			return fmt.Errorf("upload error: %w", err)
		}
	default:
		if err != nil {
			return fmt.Errorf("error sending file upload request: %w", err)
		}
	}

	// Check response
	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case 400:
			return fmt.Errorf("missing parameters")
		case 403:
			return fmt.Errorf("invalid token or IP address")
		case 409:
			return fmt.Errorf("blocked by another session")
		case 500:
			return fmt.Errorf("unknown error by receiver")
		}
		return fmt.Errorf("file upload failed: received status code %d", resp.StatusCode)
	}

	fmt.Println() // Add newline for cleaner progress bar display
	logger.Success("File uploaded successfully")
	return nil
}

// SendFile discovers devices, lets the user pick one, and sends the file
func SendFile(path string) error {
	updates := make(chan []models.SendModel)
	discovery.ListenAndStartBroadcasts(updates)
	fmt.Println("Please select a device you want to send file to:")
	ip, err := tui.SelectDevice(updates)
	if err != nil {
		return err
	}
	response, err := SendFileToOtherDevicePrepare(ip, path)
	if err != nil {
		return err
	}

	// Create a context for cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use the shared HTTP server to handle cancel requests
	logger.Info("Registering cancel handler for session: ", response.SessionID)
	RegisterCancelHandler(response.SessionID, cancel)
	defer UnregisterCancelHandler(response.SessionID)

	// Walk directory and upload files
	err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fileId := info.Name()
			token, ok := response.Files[fileId]
			if !ok {
				return fmt.Errorf("token not found for file: %s", fileId)
			}
			err = uploadFile(ctx, ip, response.SessionID, fileId, token, filePath)
			if err != nil {
				return fmt.Errorf("error uploading file: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking the path: %w", err)
	}

	return nil
}

func NormalSendHandler(w http.ResponseWriter, r *http.Request) {
	logger.Info("Handling upload request...") // Debug log - request start

	// Limit form data size (10 MB, adjustable as needed)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	// Get the uploaded directory name (from frontend hidden input)
	uploadedDirName := r.FormValue("directoryName")
	logger.Debugf("directoryName from form: '%s'\n", uploadedDirName) // Debug log - directoryName value

	// Get all uploaded files
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		http.Error(w, "No files uploaded", http.StatusBadRequest)
		return
	}

	uploadDir := config.ConfigData.SaveDir
	finalUploadDir := uploadDir // Default final upload directory

	// Only create a subdirectory if the frontend provided a non-empty directory name
	if uploadedDirName != "" {
		finalUploadDir = filepath.Join(uploadDir, uploadedDirName)
	} else {
		logger.Debug("No directoryName provided, uploading to root uploads dir.") // Debug log - no directoryName
	}
	logger.Debugf("Final upload directory: '%s'\n", finalUploadDir)

	// Create the final upload directory if it doesn't exist
	if err := os.MkdirAll(finalUploadDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create upload directory: %v", err), http.StatusInternalServerError)
		return
	}

	// Iterate over all files and save them
	for _, fileHeader := range files {
		// Open the uploaded file
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %v", err), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Build destination path (using finalUploadDir as root)
		destPath := filepath.Join(finalUploadDir, fileHeader.Filename)
		logger.Infof("Saving file '%s' to destPath: '%s'\n", fileHeader.Filename, destPath) // Debug log - file dest path

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			http.Error(w, fmt.Sprintf("Failed to create directory: %v", err), http.StatusInternalServerError)
			return
		}

		// Create destination file
		dst, err := os.Create(destPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Write uploaded file content to destination file
		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save file: %v", err), http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Upload successful: %d file(s) saved to %s\n", len(files), finalUploadDir)
}
