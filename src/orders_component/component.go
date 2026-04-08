package orders_component

import "github.com/kirilltahmazidi/aggregator/internal/models"

const Topic = "components.agregator.orders"

var Actions = []models.MessageType{
	models.MsgCreateOrder,
	models.MsgSelectExecutor,
	models.MsgAutoSearchExecutor,
}

func Handles(action models.MessageType) bool {
	for _, candidate := range Actions {
		if candidate == action {
			return true
		}
	}
	return false
}
