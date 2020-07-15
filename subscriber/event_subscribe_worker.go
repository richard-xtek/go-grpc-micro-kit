package subscriber

var (
	subscriberWorkers *SubscriberWorker
)

// Subscriber ...
type Subscriber interface {
	Start() error
	Stop() error
}

// SubscriberWorker ...
type SubscriberWorker struct {
	workerRegistries []Subscriber
}

// GetSubscriberWorker ...
func GetSubscriberWorker() *SubscriberWorker {
	if subscriberWorkers == nil {
		subscriberWorkers = &SubscriberWorker{}
	}
	return subscriberWorkers
}

// Register ...
func (w *SubscriberWorker) Register(subscribers ...Subscriber) {
	for _, subscriber := range subscribers {
		w.workerRegistries = append(w.workerRegistries, subscriber)
	}
}

// Start ...
func (w *SubscriberWorker) Start() error {
	for _, subscriber := range w.workerRegistries {
		if err := subscriber.Start(); err != nil {
			return err
		}
	}
	return nil
}

// Stop ...
func (w *SubscriberWorker) Stop() error {
	for _, subscriber := range w.workerRegistries {
		if err := subscriber.Stop(); err != nil {
			return err
		}
	}
	return nil
}
