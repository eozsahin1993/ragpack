package meta

type MetaStore interface {
	CollectionStore
	JobStore
	DocumentStore
	APIKeyStore
	Close() error
}
