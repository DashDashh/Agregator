package contracts_component

import "github.com/kirilltahmazidi/aggregator/internal/models"

const Topic = "components.agregator.contracts"

var Actions = []models.MessageType{
	models.MsgConcludeContract,
	models.MsgConfirmExecution,
	models.MsgCreateDispute,
}

func Handles(action models.MessageType) bool {
	for _, candidate := range Actions {
		if candidate == action {
			return true
		}
	}
	return false
}
