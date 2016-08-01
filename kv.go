package main

import ()

type KvBackend interface {
	Read(key string, recursive bool) (*string, error)
	Write(key string, value string, ttl string) error
}
