package main

import (
	"crypto/rand"
	"fmt"
)

type Server interface {
	Serve() error
}

type Application struct {
	ProductManager ProductManager
	ClientManager  ClientManager
	Server         Server
}

func NewID() string {
	b := make([]byte, 16)
	rand.Read(b)

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
