package storage

type BridgeStorage interface {
	Read(key string) ([]byte, error)
	Write(key string, data []byte) error
	Delete(keys []string) error
	List() ([]string, error)
}