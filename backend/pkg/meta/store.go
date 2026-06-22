package meta

type MetaStore interface {
	CollectionStore
	JobStore
	Close() error
}
