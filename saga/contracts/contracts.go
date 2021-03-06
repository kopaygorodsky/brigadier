package contracts

import (
	"github.com/go-foreman/foreman/pubsub/message"
	"github.com/go-foreman/foreman/runtime/scheme"
)

const (
	systemGroup scheme.Group = "systemSaga"
)

// RegisterSagaContracts registers all system saga contacts in specified scheme
func RegisterSagaContracts(scheme scheme.KnownTypesRegistry) {
	scheme.AddKnownTypes(
		systemGroup,
		&StartSagaCommand{},
		&RecoverSagaCommand{},
		&CompensateSagaCommand{},
		&SagaCompletedEvent{},
		&SagaChildCompletedEvent{},
	)
}

// StartSagaCommand once received will create SagaInstance, save it to Store and Start()
type StartSagaCommand struct {
	message.ObjectMeta
	SagaUID   string         `json:"saga_uid"`
	ParentUID string         `json:"parent_uid"`
	Saga      message.Object `json:"saga"`
}

type RecoverSagaCommand struct {
	message.ObjectMeta
	SagaUID string `json:"saga_uid"`
}

type CompensateSagaCommand struct {
	message.ObjectMeta
	SagaUID string `json:"saga_uid"`
}

type SagaCompletedEvent struct {
	message.ObjectMeta
	SagaUID string `json:"saga_uid"`
}

type SagaChildCompletedEvent struct {
	message.ObjectMeta
	SagaUID string `json:"saga_uid"`
}
