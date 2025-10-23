package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	appinterfaces "github.com/kristianrpo/document-management-microservice/internal/application/interfaces"
	"github.com/kristianrpo/document-management-microservice/internal/domain/events"
)

// DocumentDownloadHandler downloads pre-signed URLs and stores them via DocumentUploader
type DocumentDownloadHandler struct {
	uploader   appinterfaces.DocumentUploader
	publisher  appinterfaces.MessagePublisher
	readyQueue string
}

// NewDocumentDownloadHandler creates an instance of DocumentDownloadHandler
func NewDocumentDownloadHandler(uploader appinterfaces.DocumentUploader, publisher appinterfaces.MessagePublisher, readyQueue string) *DocumentDownloadHandler {
	return &DocumentDownloadHandler{
		uploader:   uploader,
		publisher:  publisher,
		readyQueue: readyQueue,
	}
}

// HandleDownloadRequested processes the event, downloads URLs and uploads them, then publishes a DocumentsReadyEvent
func (h *DocumentDownloadHandler) HandleDownloadRequested(ctx context.Context, message []byte) error {
	var evt events.DocumentDownloadRequestedEvent
	if err := json.Unmarshal(message, &evt); err != nil {
		log.Printf("failed to unmarshal document download requested event: %v", err)
		return fmt.Errorf("unmarshal error: %w", err)
	}

	log.Printf("processing download requested for citizen=%d urls=%d", evt.IDCitizen, len(evt.URLs))

	client := &http.Client{}
	success := true
	msg := ""

	for _, u := range evt.URLs {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			log.Printf("failed to build request for url %s: %v", u, err)
			success = false
			msg = err.Error()
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("failed to download url %s: %v", u, err)
			success = false
			msg = err.Error()
			continue
		}
		if resp.Body == nil {
			log.Printf("empty body for url %s", u)
			success = false
			msg = "empty body"
			_ = resp.Body.Close()
			continue
		}

		// We need an io.ReadSeeker; copy to buffer
		data, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			log.Printf("failed reading response for url %s: %v", u, err)
			success = false
			msg = err.Error()
			continue
		}

		r := io.NewSectionReader(bytes.NewReader(data), 0, int64(len(data)))

		// Use filename placeholder since pre-signed URL may not contain filename
		filename := "downloaded-file"
		if _, err := h.uploader.UploadFromReader(ctx, r, filename, int64(len(data)), evt.IDCitizen); err != nil {
			log.Printf("failed uploading downloaded file for url %s: %v", u, err)
			success = false
			msg = err.Error()
			continue
		}
	}

	status := "success"
	if !success {
		status = "failure"
	}

	readyEvent := events.DocumentsReadyEvent{
		IDCitizen: evt.IDCitizen,
		Status:    status,
		Message:   msg,
	}

	payload, err := json.Marshal(readyEvent)
	if err != nil {
		log.Printf("failed to marshal ready event: %v", err)
		return fmt.Errorf("marshal error: %w", err)
	}

	if err := h.publisher.Publish(ctx, h.readyQueue, payload); err != nil {
		log.Printf("failed to publish documents ready event: %v", err)
		return fmt.Errorf("publish error: %w", err)
	}

	return nil
}
