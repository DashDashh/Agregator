package handler

import (
	"log"

	"github.com/kirilltahmazidi/aggregator/internal/models"
	"github.com/kirilltahmazidi/aggregator/internal/response"
)

func errResponse(req models.Request, msg string) models.Response {
	log.Printf("[handler] error correlation_id=%s: %s", req.GetCorrelationID(), msg)
	return response.Err(req, msg)
}
