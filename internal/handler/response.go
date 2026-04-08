package handler

import (
	"log"
	"time"

	"github.com/kirilltahmazidi/aggregator/internal/models"
)

func errResponse(req models.Request, msg string) models.Response {
	log.Printf("[handler] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return models.Response{
		Action:        models.ResponseAction,
		Sender:        models.DefaultSender,
		CorrelationID: req.GetCorrelationID(),
		Success:       false,
		Error:         msg,
		Timestamp:     time.Now().UTC().Format(time.RFC3339Nano),
	}
}
