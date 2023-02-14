/*
Copyright 2023 k0s authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/k0sproject/k0s/pkg/apis/k0s.k0sproject.io/v1beta1"
	"github.com/k0sproject/k0s/pkg/constant"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	initInterval  = 5 * time.Second
	checkInterval = 15 * time.Second
)

type HealthCheck struct {
	etcdConf  *v1beta1.EtcdConfig
	k0sVars   constant.CfgVars
	ticker    *time.Ticker
	tlsConfig *tls.Config
	members   []*etcdserverpb.Member
}

// NewHealthCheck creates a new etcd health checker and starts the check loop
func NewHealthCheck(etcdConf *v1beta1.EtcdConfig, k0sVars constant.CfgVars) *HealthCheck {

	ticker := time.NewTicker(checkInterval)
	hc := &HealthCheck{
		etcdConf: etcdConf,
		ticker:   ticker,
		k0sVars:  k0sVars,
	}

	go hc.Init()
	return hc
}

func (h *HealthCheck) Init() {
	err := h.getTLSConfig(h.etcdConf, h.k0sVars)
	if err != nil {
		logrus.Error("Unable to parse etcd TLS config. In theory this can't happen.", err)
		return
	}
	fmt.Printf("VALADAS: TLS %+v\n", h.tlsConfig)
	h.waitForClient()

}

func (h *HealthCheck) waitForClient() {
	for {
		cli, err := h.newClientV3(nil)
		if err != nil {
			fmt.Printf("VALADAS: get client err: %v\n", err)
			time.Sleep(initInterval)
			continue
		}
		fmt.Printf("VALADAS: client: %#v\n", cli)
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
		defer cancel()
		resp, err := cli.MemberList(ctx)
		if err != nil {
			fmt.Printf("VALADAS: Member list err: %v\n", err)
			time.Sleep(initInterval)
			continue
		}
		fmt.Printf("VALADAS: client works %+v\n", resp)
		h.members = resp.Members

	}

}
func (h *HealthCheck) getTLSConfig(etcdConf *v1beta1.EtcdConfig, k0sVars constant.CfgVars) error {
	if etcdConf.IsTLSEnabled() {
		fmt.Printf("VALADAS k0sVars: %+v\n", k0sVars)
		tlsInfo := transport.TLSInfo{
			CertFile:      etcdConf.GetCertFilePath(k0sVars.CertRootDir),
			KeyFile:       etcdConf.GetKeyFilePath(k0sVars.CertRootDir),
			TrustedCAFile: etcdConf.GetCaFilePath(k0sVars.CertRootDir),
		}
		fmt.Printf("VALADAS tlsInfo: %+v\n", tlsInfo)
		t, err := tlsInfo.ClientConfig()
		if err != nil {
			return err
		}
		h.tlsConfig = t
	}
	return nil
}

func (h *HealthCheck) newClientV3(endpoints []string) (*clientv3.Client, error) {
	if endpoints == nil {
		endpoints = h.etcdConf.GetEndpoints()
	}
	cfg := clientv3.Config{
		Endpoints: endpoints,
		TLS:       h.tlsConfig,
	}
	return clientv3.New(cfg)
}

func (h *HealthCheck) checkLoop(ticker <-chan time.Time) {
	for range ticker {
		continue
	}
}

// HealthCheck summarizes the health of the etcd cluster in a single error
// message. If everything is OK returns nil
func (h *HealthCheck) Healthy() error {
	return h.isMemberHealthy()
}

func (h *HealthCheck) isMemberHealthy() error {
	fmt.Println("VALADAS: called")
	cli, err := h.newClientV3(nil)
	if err != nil {
		return nil
	}
	fmt.Printf("VALADAS: client err: %v\n", err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10*time.Second))
	defer cancel()
	resp, err := cli.Get(ctx, "health")
	fmt.Printf("VALADAS: health resp: %v\n", resp)
	fmt.Printf("VALADAS: health err: %v\n", err)
	return nil

}

func (h *HealthCheck) Stop() {
	h.ticker.Stop()
}
