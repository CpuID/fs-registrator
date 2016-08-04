package main

import (
	"errors"
	"fmt"
	etcd_client "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"strings"
	"time"
)

type KvBackendEtcd struct {
	Kapi   etcd_client.KeysAPI
	Prefix string
}

func NewKvBackendEtcd(conf map[string]string) (KvBackend, error) {
	for _, v := range []string{"host", "port", "prefix"} {
		if _, ok := conf[v]; ok == false {
			return nil, errors.New(fmt.Sprintf("etcd: '%s' key does not exist in conf.", v))
		}
	}
	cfg := etcd_client.Config{
		// TODO: do we want to specify multiple etcd hosts?
		Endpoints: []string{fmt.Sprintf("http://%s:%s", conf["host"], conf["port"])},
		Transport: etcd_client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := etcd_client.New(cfg)
	if err != nil {
		return nil, err
	}
	return &KvBackendEtcd{
		Kapi:   etcd_client.NewKeysAPI(c),
		Prefix: conf["prefix"],
	}, nil
}

func (k *KvBackendEtcd) BackendName() string {
	return "etcd"
}

func (k *KvBackendEtcd) GetPrefix() string {
	return k.Prefix
}

func (k *KvBackendEtcd) UseKey(key string) string {
	use_key := k.Prefix
	if len(key) > 0 {
		use_key = fmt.Sprintf("%s/%s", use_key, key)
	}
	return use_key
}

// If the key is a prefix (recursive lookup), set recursive = true
// Results will be key/value in a map.
func (k *KvBackendEtcd) Read(key string, recursive bool) (*map[string]string, error) {
	use_key := k.UseKey(key)
	log.Printf("etcd.Read(): Getting '%s' key value (recursive: %t)", use_key, recursive)
	// TODO: parse option for recursive to .Get()
	resp, err := k.Kapi.Get(context.Background(), use_key, nil)
	var results map[string]string
	if err != nil {
		if strings.Contains(err.Error(), "100: Key not found") {
			return &results, errors.New("KEY_NOT_FOUND")
		} else {
			return &results, err
		}
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
		log.Printf("Count of child nodes: %d\n", len(resp.Node.Nodes))
	}
	// TODO: parse out etcd_client.Node, get a string value?
	return &results, nil
}

func (k *KvBackendEtcd) Write(key string, value string, ttl int) error {
	use_key := k.UseKey(key)
	log.Printf("etcd.Write(): Writing '%s' key value", use_key)
	resp, err := k.Kapi.Set(context.Background(), use_key, value, nil)
	if err != nil {
		return err
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}
	return nil
}

func (k *KvBackendEtcd) Delete(key string) error {
	use_key := k.UseKey(key)
	log.Printf("etcd.Delete(): Deleting '%s' key value", use_key)
	resp, err := k.Kapi.Delete(context.Background(), use_key, nil)
	if err != nil {
		return err
	} else {
		// print common key info
		log.Printf("Delete is done. Metadata is %q\n", resp)
	}
	return nil
}
