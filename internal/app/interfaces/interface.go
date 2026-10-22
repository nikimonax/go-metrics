// Пакет interfaces определяет интерфейсы application слоя.
//
// Конкретные реализации находятся во внешних слоях.

package interfaces

import (
	"context"

	"github.com/nikimonax/go-metrics/internal/domain"
)

// Интерфейс хранилища метрик.
// Предоставляет операции обновления, чтения метрик.
type MetricRepository interface {
	// Обновляет одну метрику в хранилище.
	Update(context.Context, domain.Metric) error
	// Обновляет пакет метрик в хранилище.
	UpdateBatch(context.Context, []domain.Metric) error
	// Возвращает метрику по её типу и имени.
	Get(context.Context, domain.MetricType, domain.MetricName) (domain.Metric, error)
	// Возвращает все метрики из хранилища.
	GetAll(context.Context) ([]domain.Metric, error)
	// Возвращает и удаляет все метрики из хранилища.
	PopAll(context.Context) ([]domain.Metric, error)
	// Удаляет все метрики из хранилища.
	Clear(context.Context) error
}

// Интерфейс сборщика метрик.
// Отвечает за сбор текущих значений метрик с устройства/процесса.
type MetricCollector interface {
	// Собирает и возвращает список метрик.
	Collect() ([]domain.Metric, error)
}

// Интерфейс клиента для отправки метрик на внешние сервисы.
type MetricGateway interface {
	// Отправляет одну метрику.
	Send(context.Context, domain.Metric) error
	// Отправляет пакет метрик.
	SendBatch(context.Context, []domain.Metric) error
}

// Интерфейс выгрузки/загрузки метрик в файл.
// Используется для сохранения метрик на диск и их восстановления.
type MetricDumper interface {
	// Сохраняет переданные метрики на диск.
	Save([]domain.Metric) error
	// Загружает метрики с диска.
	Load() ([]domain.Metric, error)
}

// Интерфейс события, публикуемого через EventPublisher.
type Event interface {
	// Возвращает имя события.
	Name() string
}

// Интерфейс издателя событий.
type EventPublisher interface {
	// Публикует переданное событие.
	Publish(context.Context, Event)
}
