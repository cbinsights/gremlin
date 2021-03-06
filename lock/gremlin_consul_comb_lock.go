package lock

import (
	"fmt"

	consulapi "github.com/hashicorp/consul/api"
)

// This lock is a dual implementation of the Local Lock and the Consul Lock
// It is designed to be implemented on a distributed system, with less strain on Consul than the pure Consul lock

type ConsulCombinationLockClient struct {
	ConsulClient *ConsulLockClient
	LocalClient  *LocalLockClient
}

type ConsulCombinationLock struct {
	ConsulLock ConsulLock
	LocalLock  LocalLock
}

func NewConsulCombinationLockClient(consulAddress, consulBaseFolder string, consulLockOptions *consulapi.LockOptions) (*ConsulCombinationLockClient, error) {
	consulClient, err := NewConsulLockClient(consulAddress, consulBaseFolder, consulLockOptions)
	if err != nil {
		return nil, err
	}
	localClient := NewLocalLockClient()
	return &ConsulCombinationLockClient{
		ConsulClient: &consulClient,
		LocalClient:  localClient,
	}, nil
}

func (c ConsulCombinationLockClient) LockKey(key string) (Lock_i, error) {
	consulLock, err := c.ConsulClient.LockKey(key)
	if err != nil {
		return nil, err
	}
	localLock, err := c.LocalClient.LockKey(key)
	if err != nil {
		return nil, err
	}
	return ConsulCombinationLock{
		ConsulLock: consulLock.(ConsulLock),
		LocalLock:  localLock.(LocalLock),
	}, nil
}

func (lock ConsulCombinationLock) Lock() error {
	err := lock.LocalLock.Lock()
	if err != nil {
		return err
	}
	err = lock.ConsulLock.Lock()
	return err
}

func (lock ConsulCombinationLock) Unlock() error {
	errResponse := ""
	err := lock.ConsulLock.Unlock()
	if err != nil {
		errResponse = fmt.Sprintf("Consul unlock error: %v", err)
	}
	err = lock.LocalLock.Unlock()
	if err != nil {
		errResponse = fmt.Sprintf("%s. Local unlock error: %v", errResponse, err)
	}
	if errResponse != "" {
		return fmt.Errorf(errResponse)
	}
	return nil
}
