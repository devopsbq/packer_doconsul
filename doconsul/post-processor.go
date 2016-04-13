package doconsul

import (
	"fmt"
	"log"

	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
)

const (
	digitalOceanBuilderID = "pearkes.digitalocean"
	consulPrefixKey       = "packer/doconsul/"
)

// Config contains configuration specific to this post-processor.
type Config struct {
	common.PackerConfig `mapstructure:",squash"`

	// Consul fields
	ConsulAddress string `mapstructure:"consul_address"`
	ConsulScheme  string `mapstructure:"consul_scheme"`
	ConsulToken   string `mapstructure:"consul_token"`

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

	// required configuration
	templates := map[string]*string{
		"snapshot_name": &p.config.SnapshotName,
	}

	// verifying configuration is set
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

	log.Printf("Creating consul client")
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

	p.client, err = api.NewClient(consulConfig)
	if err != nil {
		return a, false, err
	}

	key := p.config.SnapshotName

	if p.config.SnapshotVersion != "" {
		key = fmt.Sprintf("%s-%s", key, p.config.SnapshotVersion)
	}

	kvpair := api.KVPair{Key: fmt.Sprintf("%s%s", consulPrefixKey, key), Value: []byte(snapshotID)}
	log.Printf(fmt.Sprintf("Putting key %s with value %s in consul...", key, snapshotID))
	ui.Message(fmt.Sprintf("Putting key %s with value %s in consul...", key, snapshotID))
	if _, err = p.client.KV().Put(&kvpair, nil); err != nil {
		return a, false, err
	}

	return a, true, nil
}

// TODO: godep?

// TODO: goxc?
