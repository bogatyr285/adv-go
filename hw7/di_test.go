package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v di_test.go

type UserService struct {
	// not need to implement
	NotEmptyStruct bool
}
type MessageService struct {
	// not need to implement
	NotEmptyStruct bool
}
type SingletonService struct {
	// not need to implement
	NotEmptyStruct bool
}

type Container struct {
	dependencies map[string]any
}

func NewContainer() *Container {
	return &Container{
		dependencies: make(map[string]any),
	}
}

func (c *Container) RegisterType(name string, constructor any) {
	if _, ok := c.dependencies[name]; ok {
		return
	}

	c.dependencies[name] = constructor
}

func (c *Container) RegisterSingletonType(name string, fn any) {
	if _, ok := c.dependencies[name]; ok {
		return
	}

	constructor, ok := fn.(func() any)
	if !ok {
		return
	}

	c.dependencies[name] = constructor()
}

func (c *Container) Resolve(name string) (any, error) {
	fn, ok := c.dependencies[name]
	if !ok {
		return nil, fmt.Errorf("%s is not exist", name)
	}

	switch fn := fn.(type) {
	case func() any:
		return fn(), nil
	default:
		return fn, nil
	}
}

func TestDIContainer(t *testing.T) {
	container := NewContainer()
	container.RegisterType("UserService", func() any {
		return &UserService{}
	})
	container.RegisterType("MessageService", func() any {
		return &MessageService{}
	})
	container.RegisterSingletonType("SingletonService", func() any {
		return &SingletonService{}
	})

	userService1, err := container.Resolve("UserService")
	assert.NoError(t, err)
	userService2, err := container.Resolve("UserService")
	assert.NoError(t, err)

	u1 := userService1.(*UserService)
	u2 := userService2.(*UserService)
	assert.False(t, u1 == u2)

	messageService, err := container.Resolve("MessageService")
	assert.NoError(t, err)
	assert.NotNil(t, messageService)

	paymentService, err := container.Resolve("PaymentService")
	assert.Error(t, err)
	assert.Nil(t, paymentService)

	singletonService1, err := container.Resolve("SingletonService")
	assert.NoError(t, err)
	singletonService2, err := container.Resolve("SingletonService")
	assert.NoError(t, err)

	s1 := singletonService1.(*SingletonService)
	s2 := singletonService2.(*SingletonService)
	assert.True(t, s1 == s2)
}
