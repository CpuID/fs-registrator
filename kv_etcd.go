package main

import (
	"fmt"
	etcd_client "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"time"
)

type KvBackendEtcd struct {
	Kapi etcd_client.KeysAPI
}

func (k KvBackendEtcd) SetupEtcdClient(host string, port uint) error {
	cfg := etcd_client.Config{
		// TODO: do we want to specify multiple etcd hosts?
		Endpoints: []string{fmt.Sprintf("http://%s:%d", host, port)},
		Transport: etcd_client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := etcd_client.New(cfg)
	if err != nil {
		return err
	}
	k.Kapi = etcd_client.NewKeysAPI(c)
	return nil
}

// If the key is a prefix (recursive lookup), set recursive = true
func (k KvBackendEtcd) Read(key string, recursive bool) (*string, error) {
	log.Printf("readKv(): Getting '%s' key value (recursive: %t)", key, recursive)
	// TODO: parse option for recursive to .Get()
	resp, err := k.Kapi.Get(context.Background(), key, nil)
	if err != nil {
		return new(string), err
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
	// TODO: parse out etcd_client.Node, get a string value?
	return new(string), nil
}

func (k KvBackendEtcd) Write(key string, value string, ttl string) error {
	log.Printf("writeKv(): Writing '%s' key value", key)
	resp, err := k.Kapi.Set(context.Background(), key, value, nil)
	if err != nil {
		return err
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}
	return nil
}
