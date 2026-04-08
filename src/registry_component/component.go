package registry_component

import "github.com/kirilltahmazidi/aggregator/internal/models"

const Topic = "components.agregator.registry"

var Actions = []models.MessageType{
	models.MsgRegisterOperator,
	models.MsgRegisterCustomer,
}

func Handles(action models.MessageType) bool {
	for _, candidate := range Actions {
		if candidate == action {
			return true
		}
	}
	return false
}
