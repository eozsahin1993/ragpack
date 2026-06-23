package meta

type MetaStore interface {
	CollectionStore
	JobStore
	DocumentStore
	Close() error
}
