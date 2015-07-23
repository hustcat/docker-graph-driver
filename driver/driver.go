package graphdriver

import (
	"errors"
	"fmt"
	"os"
	"path"
)

type InitFunc func(root string, options []string) (Driver, error)

// Driver represent the interface a driver must fulfill.
type Driver interface {
	Create(id, parent string) error
	Remove(id string) error
	Get(id, mountLabel string) (string, error)
	Put(id string) error
	Exists(id string) bool
	Status() [][2]string
	GetMetadata(id string) (map[string]string, error)
	Cleanup() error
}

var (
	DefaultDriver string
	// All registred drivers
	drivers map[string]InitFunc
	// Slice of drivers that should be used in an order
	priority = []string{
		"vfs",
		"rbd",
	}

	ErrNotSupported   = errors.New("driver not supported")
	ErrPrerequisites  = errors.New("prerequisites for driver not satisfied (wrong filesystem?)")
	ErrIncompatibleFS = fmt.Errorf("backing file system is unsupported for this graph driver")
)

func init() {
	drivers = make(map[string]InitFunc)
}

func Register(name string, initFunc InitFunc) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Name already registered %s", name)
	}
	drivers[name] = initFunc

	return nil
}

func GetDriver(name, home string, options []string) (Driver, error) {
	if initFunc, exists := drivers[name]; exists {
		return initFunc(path.Join(home, name), options)
	}
	return nil, ErrNotSupported
}

func New(root string, options []string) (driver Driver, err error) {
	for _, name := range []string{os.Getenv("DOCKER_DRIVER"), DefaultDriver} {
		if name != "" {
			return GetDriver(name, root, options)
		}
	}

	// Check for priority drivers first
	for _, name := range priority {
		driver, err = GetDriver(name, root, options)
		if err != nil {
			if err == ErrNotSupported || err == ErrPrerequisites || err == ErrIncompatibleFS {
				continue
			}
			return nil, err
		}
		return driver, nil
	}

	// Check all registered drivers if no priority driver is found
	for _, initFunc := range drivers {
		if driver, err = initFunc(root, options); err != nil {
			if err == ErrNotSupported || err == ErrPrerequisites || err == ErrIncompatibleFS {
				continue
			}
			return nil, err
		}
		return driver, nil
	}
	return nil, fmt.Errorf("No supported storage backend found")
}
