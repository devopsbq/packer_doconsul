package doconsul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	digitalOceanBuilderID = "pearkes.digitalocean"
	consulPrefixKey       = "snaps/do/"
)

// Config contains configuration specific to this post-processor.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Consul fields
	ConsulAddress string `mapstructure:"consul_address"`
	ConsulScheme  string `mapstructure:"consul_scheme"`
	ConsulToken   string `mapstructure:"consul_token"`

	// Experimental TLS support
	CAFile         string `mapstructure:"ca_file"`
	CertFile       string `mapstructure:"cert_file"`
	KeyFile        string `mapstructure:"key_file"`
	SkipTLSVerifiy bool   `mapstructure:"skip_tls_verify"`

	// Image info fields, which will be stored in Consul
	SnapshotName    string `mapstructure:"snapshot_name"` // Required
	SnapshotVersion string `mapstructure:"snapshot_version"`

	ctx interpolate.Context
}

// PostProcessor is the struct which contains the configuration and the Consul client
type PostProcessor struct {
	config Config
	client *api.Client
}

// Configure method implementation
func (p *PostProcessor) Configure(raws ...interface{}) error {
	log.Printf("Configuring Post Processor with content: %v", raws)
	excludeInterpolate := []string{}
	opts := config.DecodeOpts{Interpolate: true, InterpolateContext: &p.config.ctx, InterpolateFilter: &interpolate.RenderFilter{Exclude: excludeInterpolate}}
	if err := config.Decode(&p.config, &opts, raws...); err != nil {
		return err
	}

	errs := new(packer.MultiError)

	if p.config.ConsulAddress != "" {
		consulAddr, err := parseConsulAddress(p.config.ConsulAddress)
		if err != nil {
			errs = packer.MultiErrorAppend(err, errs)
		} else {
			p.config.ConsulAddress = consulAddr
		}
	}

	if p.config.ConsulScheme != "" && p.config.ConsulScheme != "http" && p.config.ConsulScheme != "https" {
		errs = packer.MultiErrorAppend(fmt.Errorf("Invalid Consul scheme: %s", p.config.ConsulScheme), errs)
	}

	// required configuration
	templates := map[string]*string{
		"snapshot_name": &p.config.SnapshotName,
	}

	// if any of the TLS certificates is set, the others also must be set.
	if p.config.CAFile != "" || p.config.CertFile != "" || p.config.KeyFile != "" {
		templates["ca_file"] = &p.config.CAFile
		templates["cert_file"] = &p.config.CertFile
		templates["key_file"] = &p.config.KeyFile
	}

	// verifying configuration is set
	log.Printf("Fields to check: %v", templates)
	for key, value := range templates {
		if *value == "" {
			e := fmt.Errorf("%s must be set", key)
			log.Printf("Error: %s", e.Error())
			errs = packer.MultiErrorAppend(e, errs)
		}
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

// PostProcess method implementation
func (p *PostProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	log.Printf("Post Processing artifact: %v", a)
	if a.BuilderId() != digitalOceanBuilderID {
		return nil, false, fmt.Errorf("Unknown artifact type: %s", a.BuilderId())
	}

	snapshotID, err := getImageIDfromDOArtifact(a)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return nil, false, err
	}

	consulConfig := api.DefaultConfig()
	if p.config.ConsulAddress != "" {
		consulConfig.Address = p.config.ConsulAddress
	}

	if p.config.ConsulScheme != "" {
		consulConfig.Scheme = p.config.ConsulScheme
	}

	if p.config.ConsulToken != "" {
		consulConfig.Token = p.config.ConsulToken
	}

	if p.config.CAFile != "" {
		p.config.ConsulScheme = "https"
		apiTLSConfig := &api.TLSConfig{
			Address:  p.config.ConsulAddress,
			CAFile:   p.config.CAFile,
			CertFile: p.config.CertFile,
			KeyFile:  p.config.KeyFile,
		}

		transport := cleanhttp.DefaultPooledTransport()
		if transport.TLSClientConfig, err = api.SetupTLSConfig(apiTLSConfig); err != nil {
			return nil, false, err
		}

		transport.TLSClientConfig.InsecureSkipVerify = p.config.SkipTLSVerifiy
		consulConfig.HttpClient.Transport = transport
	}

	log.Printf("Creating consul client with config: %v", consulConfig)
	p.client, err = api.NewClient(consulConfig)
	if err != nil {
		return a, false, err
	}

	key := "latest"

	if p.config.SnapshotVersion != "" {
		key = p.config.SnapshotVersion
	}

	kvpair := api.KVPair{Key: fmt.Sprintf("%s%s/%s", consulPrefixKey, p.config.SnapshotName, key), Value: []byte(snapshotID)}
	log.Printf(fmt.Sprintf("Putting key %s%s/%s with value %s in Consul...", consulPrefixKey, p.config.SnapshotName, key, snapshotID))
	ui.Message(fmt.Sprintf("Putting key %s%s/%s with value %s in Consul...", consulPrefixKey, p.config.SnapshotName, key, snapshotID))
	if _, err = p.client.KV().Put(&kvpair, nil); err != nil {
		return a, false, err
	}

	return a, true, nil
}

// TODO: goxc?
