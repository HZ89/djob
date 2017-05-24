package djob

import "github.com/docker/libkv/store"

type Store struct {
	Client store.Store
	keyspace string
	backend string
}
