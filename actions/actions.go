package actions

import (
	"encoding/json"

	"github.com/CenturyLinkLabs/prettycli"
	"github.com/CenturyLinkLabs/zodiac/cluster"
	"github.com/CenturyLinkLabs/zodiac/composer"
	"github.com/CenturyLinkLabs/zodiac/proxy"
	"github.com/samalba/dockerclient"
)

const ProxyAddress = "localhost:31981"

var (
	DefaultProxy    proxy.Proxy
	DefaultComposer composer.Composer
)

func init() {
	DefaultProxy = proxy.NewHTTPProxy(ProxyAddress)
	DefaultComposer = composer.NewExecComposer(ProxyAddress)
}

type Options struct {
	Args  []string
	Flags map[string]string
}

type Zodiaction func(cluster.Cluster, Options) (prettycli.Output, error)

type DeploymentManifests []DeploymentManifest

type DeploymentManifest struct {
	Services   []Service
	DeployedAt string
}

type Service struct {
	Name            string
	ContainerConfig dockerclient.ContainerConfig
}

func collectRequests(endpoint cluster.Endpoint, args []string) []cluster.ContainerRequest {
	// TODO: handle error
	go DefaultProxy.Serve(endpoint)
	// TODO: handle error
	defer DefaultProxy.Stop()

	// TODO: handle error
	// TODO: args not passed!
	DefaultComposer.Run(args)
	return DefaultProxy.DrainRequests()
}

func startServices(services []Service, manifests DeploymentManifests, endpoint cluster.Endpoint) error {
	manifestsBlob, err := json.Marshal(manifests)
	if err != nil {
		return err
	}

	for _, svc := range services {
		if svc.ContainerConfig.Labels == nil {
			svc.ContainerConfig.Labels = make(map[string]string)
		}
		svc.ContainerConfig.Labels["zodiacManifest"] = string(manifestsBlob)

		endpoint.StartContainer(svc.Name, svc.ContainerConfig)
	}

	return nil
}
