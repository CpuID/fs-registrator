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
			return nil, fmt.Errorf("etcd: '%s' key does not exist in conf.", v)
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

// If the key is a prefix (recursive lookup), set recursive = true
// Results will be key/value in a map.
func (k *KvBackendEtcd) Read(key string, recursive bool) (*map[string]string, error) {
	use_key := getKvKeyWithPrefix(k.Prefix, key)
	log.Printf("etcd.Read(): Getting '%s' key value (recursive: %t)", use_key, recursive)
	var get_options etcd_client.GetOptions
	if recursive == true {
		get_options.Recursive = true
	}
	resp, err := k.Kapi.Get(context.Background(), use_key, &get_options)
	results := make(map[string]string)
	if err != nil {
		if strings.Contains(err.Error(), "100: Key not found") {
			return &results, errors.New("KEY_NOT_FOUND")
		} else {
			return &results, err
		}
	}
	//log.Printf("Get is done. Metadata is %q\n", resp)
	//log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	//log.Printf("Count of child nodes: %d\n", len(resp.Node.Nodes))
	if resp.Node.Dir == true {
		for _, v := range resp.Node.Nodes {
			// We only support a single layer of keys under a single parent directory currently, as opposed to recursive keys.
			// Can support more layers in future as required (using a separate function call), this use case doesn't require it.
			if v.Dir == true {
				return new(map[string]string), errors.New("UNSUPPORTED_CHILD_KEY_AS_DIRECTORY")
			}
			results[stripKvKeyPrefix(k.Prefix, v.Key)] = v.Value
		}
	} else {
		result_key := stripKvKeyPrefix(k.Prefix, resp.Node.Key)
		if len(result_key) == 0 {
			// If we strip the prefix, there would be no key at all. Leave it in place, just remove leading slash instead.
			// This use case should be rare in this app.
			if resp.Node.Key[0:1] == "/" && len(resp.Node.Key) > 1 {
				result_key = resp.Node.Key[1:]
			} else {
				result_key = resp.Node.Key
			}
		}
		results[result_key] = resp.Node.Value
	}
	return &results, nil
}

func (k *KvBackendEtcd) Write(key string, value string, ttl int) error {
	use_key := getKvKeyWithPrefix(k.Prefix, key)
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
	use_key := getKvKeyWithPrefix(k.Prefix, key)
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
