package meta

type MetaStore interface {
	CollectionStore
	JobStore
	DocumentStore
	APIKeyStore
	PromptStore
	Close() error
}
