package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"status": true})
}

func VideoUploadHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	r.Body = http.MaxBytesReader(w, r.Body, 100<<20)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("video")
	if err != nil {
		http.Error(w, "Error retrieving file: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Printf("Upload: file size reported as %d bytes\n", header.Size)

	fileType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(fileType, "video/") {
		http.Error(w, "File is not a video", http.StatusBadRequest)
		return
	}

	videoID := uuid.New().String()

	videoDir := filepath.Join("storage", "videos", videoID)

	if err := os.MkdirAll(videoDir, 0755); err != nil {
		http.Error(w, "Failed to create directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	videoPath := filepath.Join(videoDir, videoID+filepath.Ext(header.Filename))
	dst, err := os.Create(videoPath)
	if err != nil {
		http.Error(w, "Failed to create file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Upload: wrote %d bytes to %s\n", written, videoPath)

	chunksDir := filepath.Join(videoDir, "chunks")
	if err := os.MkdirAll(chunksDir, 0755); err != nil {
		http.Error(w, "Failed to create chunks directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	go func() {
		if err := processVideoIntoChunks(videoPath, chunksDir); err != nil {
			fmt.Printf("Error processing video %s: %v\n", videoID, err)
		} else {
			fmt.Printf("Successfully processed video into chunks: %s\n", videoID)
			// We could send a notification via email, Slack, etc. here
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"videoId": videoID,
		"message": "Video uploaded successfully and processing started",
	})
}

// processVideoIntoChunks divides a video into streamable chunks
func processVideoIntoChunks(videoPath, chunksDir string) error {
	videoID := filepath.Base(filepath.Dir(chunksDir))

	cmd := exec.Command(
		"ffmpeg",
		"-i", videoPath,
		"-profile:v", "baseline",
		"-level", "3.0",
		"-start_number", "0",
		"-hls_time", "1",
		"-force_key_frames", "expr:gte(t,n_forced*1)",
		"-vf", "scale=-2:720",
		"-c:v", "libx264",
		"-preset", "faster",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		"-hls_list_size", "0",
		"-hls_segment_type", "mpegts",
		"-hls_flags", "independent_segments",
		"-hls_base_url", "/video/"+videoID+"/",
		"-f", "hls",
		"-hls_segment_filename", filepath.Join(chunksDir, "chunk_%03d.ts"),
		filepath.Join(chunksDir, "playlist.m3u8"),
	)
	return cmd.Run()
}

func VideoStreamHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/video/")
	components := strings.Split(path, "/")

	if len(components) == 0 || components[0] == "" {
		http.Error(w, "Missing video ID", http.StatusBadRequest)
		return
	}

	videoID := components[0]

	videoDir := filepath.Join("storage", "videos", videoID)
	if _, err := os.Stat(videoDir); os.IsNotExist(err) {
		http.Error(w, "Video not found", http.StatusNotFound)
		return
	}

	chunksDir := filepath.Join(videoDir, "chunks")
	if _, err := os.Stat(chunksDir); os.IsNotExist(err) {
		http.Error(w, "Video chunks not found", http.StatusNotFound)
		return
	}

	// Serve playlist or chunk, update url to cdn or s3 if needed
	// http.Redirect(w, r, cdnOrS3Url+videoID+"/playlist.m3u8", http.StatusFound)
	var filePath string
	if len(components) == 1 {
		filePath = filepath.Join(chunksDir, "playlist.m3u8")
	} else {
		fileName := components[len(components)-1]
		filePath = filepath.Join(chunksDir, fileName)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	switch filepath.Ext(filePath) {
	case ".m3u8":
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case ".ts":
		w.Header().Set("Content-Type", "video/mp2t")
	}

	// Allow CORS on video stream for local development purposes :)
	w.Header().Set("Access-Control-Allow-Origin", "*")

	http.ServeFile(w, r, filePath)
}
