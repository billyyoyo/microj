package config

import "github.com/billyyoyo/viper/remote"

type Manager struct {
	Store Store
}

func (c Manager) Get(key string) ([]byte, error) {
	value, err := c.Store.Get(key)
	if err != nil {
		return nil, err
	}
	return value, err
}

func (c Manager) List(key string) (remote.KVPairs, error) {
	list, err := c.Store.List(key)
	retList := make(remote.KVPairs, len(list))
	if err != nil {
		return nil, err
	}
	for i, kv := range list {
		retList[i] = &remote.KVPair{Key: kv.Key, Value: kv.Value}
	}
	return retList, err
}

func (c Manager) Set(key string, value []byte) error {
	err := c.Store.Set(key, value)
	return err
}

func (c Manager) Watch(key string, stop chan bool) <-chan *remote.Response {
	resp := make(chan *remote.Response, 0)
	backendResp := c.Store.Watch(key, stop)
	go func() {
		for {
			select {
			case <-stop:
				return
			case r := <-backendResp:
				if r.Error != nil {
					resp <- &remote.Response{nil, r.Error}
					continue
				}
				resp <- &remote.Response{r.Value, nil}
			}
		}
	}()
	return resp
}

type Store interface {
	// Get retrieves a value from a K/V Store for the provided key.
	Get(key string) ([]byte, error)

	// List retrieves all keys and values under a provided key.
	List(key string) (remote.KVPairs, error)

	// Set sets the provided key to value.
	Set(key string, value []byte) error

	// Watch monitors a K/V Store for changes to key.
	Watch(key string, stop chan bool) <-chan *remote.Response
}
