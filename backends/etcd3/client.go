package etcd3

import (
	"crypto/tls"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/pkg/transport"
	"golang.org/x/net/context"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"strings"
)

// Client is a wrapper around the etcd client
type Client struct {
	client clientv3.Client
}

// NewEtcdClient returns an *etcd.Client with a connection to named machines.
func NewEtcdClient(machines []string, cert, key, caCert string, basicAuth bool, username string, password string) (*Client, error) {
	// set tls if any one tls option set
	var cfgtls *transport.TLSInfo
	tlsinfo := transport.TLSInfo{}
	if cert != "" {
		tlsinfo.CertFile = caCert
		cfgtls = &tlsinfo
	}

	if key != "" {
		tlsinfo.KeyFile = key
		cfgtls = &tlsinfo
	}

	if caCert != "" {
		tlsinfo.CAFile = caCert
		cfgtls = &tlsinfo
	}

	cfg := clientv3.Config{
		Endpoints:               machines,
		DialTimeout: time.Duration(3) * time.Second,
	}
	if cfgtls != nil {
		clientTLS, err := cfgtls.ClientConfig()
		if err != nil {
			return nil, err
		}
		cfg.TLS = clientTLS
	}
	// if key/cert is not given but user wants secure connection, we
	// should still setup an empty tls configuration for gRPC to setup
	// secure connection.
	if cfg.TLS == nil { // FIXME && !scfg.insecureTransport {
		cfg.TLS = &tls.Config{}
	}

	if basicAuth {
		cfg.Username = username
		cfg.Password = password
	}

	c, err := clientv3.New(cfg)
	if err != nil {
		return &Client{*c}, err
	}

	return &Client{*c}, nil
}

// GetValues queries etcd for keys prefixed by prefix.
func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, key := range keys {
		opts := []clientv3.OpOption{}
		if len(key) == 0 {
			key = "\x00"
			opts = append(opts, clientv3.WithFromKey())
		} else {
			opts = append(opts, clientv3.WithPrefix())
		}

		resp, err := c.client.Get(context.Background(), key, opts...)
		if err != nil {
			return vars, err
		}

		for _, kv := range resp.Kvs {
			k, v := string(kv.Key), string(kv.Value)
			vars[k] = v
		}
	}
	return vars, nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	// return something > 0 to trigger a key retrieval from the store
	if waitIndex == 0 {
		return 1, nil
	}

	for {
		_, cancel := context.WithCancel(context.Background())
		cancelRoutine := make(chan bool)
		defer close(cancelRoutine)

		go func() {
			select {
			case <-stopChan:
				cancel()
			case <-cancelRoutine:
				return
			}
		}()

		opts := []clientv3.OpOption{clientv3.WithRev(0)}
		opts = append(opts, clientv3.WithPrefix())

		ch := c.client.Watch(context.TODO(), prefix, opts...)

		for resp := range ch {
			for _, e := range resp.Events {
				if e.Type == mvccpb.PUT {
					for _, k := range keys {
						kvk := string(e.Kv.Key)
						if strings.HasPrefix(kvk, k) {
							return uint64(e.Kv.ModRevision), nil
						}
					}

				}

			}
		}

		return 0, nil
	}
}

