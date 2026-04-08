package analytics_component

import "github.com/kirilltahmazidi/aggregator/internal/models"

const Topic = "components.agregator.analytics"

var Actions = []models.MessageType{
	models.MsgGetAnalytics,
}

func Handles(action models.MessageType) bool {
	for _, candidate := range Actions {
		if candidate == action {
			return true
		}
	}
	return false
}
