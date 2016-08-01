package main

import (
	"fmt"
	etcd_client "github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"log"
	"time"
)

func setupEtcdClient(host string, port uint8) (etcd_client.KeysAPI, error) {
	cfg := etcd_client.Config{
		// TODO: do we want to specify multiple etcd hosts?
		Endpoints: []string{fmt.Sprintf("http://%s:%d", host, port)},
		Transport: etcd_client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := etcd_client.New(cfg)
	if err != nil {
		return etcd_client.NewKeysAPI(nil), err
	}
	return etcd_client.NewKeysAPI(c), nil
}

// If the key is a prefix (recursive lookup), set prefix = true
func readKv(kapi etcd_client.KeysAPI, key string, prefix bool) {
	// get "/foo" key's value
	log.Print("Getting '/foo' key value")
	resp, err := kapi.Get(context.Background(), "/foo", nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Get is done. Metadata is %q\n", resp)
		// print value
		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
	}
}

func writeKv(kapi etcd_client.KeysAPI, key string, value string) {
	// set "/foo" key with "bar" value
	log.Print("Setting '/foo' key with 'bar' value")
	resp, err := kapi.Set(context.Background(), "/foo", "bar", nil)
	if err != nil {
		log.Fatal(err)
	} else {
		// print common key info
		log.Printf("Set is done. Metadata is %q\n", resp)
	}
}
