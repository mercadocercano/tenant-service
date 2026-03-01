package event

import (
	"context"
)

// PublishEventUseCase es la interfaz del eventbus para publicar eventos
type PublishEventUseCase interface {
	Execute(
		ctx context.Context,
		aggregateID string,
		aggregateType string,
		eventType string,
		payload []byte,
		publishedBy string,
	) error
}

// EventPublisherAdapter adapta el PublishEventUseCase del eventbus a la interfaz del dominio
type EventPublisherAdapter struct {
	publishUseCase PublishEventUseCase
}

// NewEventPublisherAdapter crea una nueva instancia del adaptador
func NewEventPublisherAdapter(publishUseCase PublishEventUseCase) *EventPublisherAdapter {
	return &EventPublisherAdapter{
		publishUseCase: publishUseCase,
	}
}

// Publish publica un evento usando el eventbus
func (a *EventPublisherAdapter) Publish(
	ctx context.Context,
	aggregateID string,
	aggregateType string,
	eventType string,
	payload []byte,
	publishedBy string,
) error {
	return a.publishUseCase.Execute(ctx, aggregateID, aggregateType, eventType, payload, publishedBy)
}
