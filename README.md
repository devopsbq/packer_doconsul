# packer_doconsul
A Packer Post Processor for storing DigitalOcean image IDs into Consul.
The implementation is based on [packer-post-processor-consul][packer-pp]'s, from [bhourigan][bhourigan].

This post-processor takes an image from [DigitalOcean's builder][dobuilder] and stores its ID in Consul, within the path `snaps/do/<snapshot_name>/<snapshot_version>`.

### Features

* TLS support for secure communication with the Consul API.

### Installation

The first thing you need to do is to build the source code and name it correctly for Packer to be able to detect it as a plugin:

```
cd packer_doconsul
go get
go build
mv packer_doconsul ~/.packer.d/plugins/packer-post-processor-doconsul
```
For more information about installing the plugin, please check [Packer's Plugins][plugins] documentation.


### Configuration

The following parameters are **required**:

* `snapshot_name`: The name you wish to give to your snapshot.

The following parameters are **optional**:

* `consul_address`: The address of your Consul deployment. By default,  `127.0.0.1:8500`
* `consul_scheme`: The Consul scheme. This can be either `http`, or `https`. By default, `http`.
* `consul_token`: The token for communication with Consul, if needed.
* `snapshot_version`: The version you wish to give to your snapshot. By default, `latest`.
* `ca_file`: A file path to a PEM-encoded certificate authority.
* `cert_file`: A file path to a PEM-encoded certificate.
* `key_file`: A the file path to a PEM-encoded private key.
* `skip_tls_verify`: Skip server-side certificate validation.

Remember that this post-processor will store your DigitalOcean's image ID at path `snaps/do/<snapshot_name>/<snapshot_version>`.

### Basic example

Here is a basic example. As you can see, you can do interpolation for computing the value of the fields.

```
"post-processors": [
  {
    "type": "doconsul",
    "snapshot_name": "{{user `snap_name`}}",
    "snapshot_version": "0.1.0",
    "consul_address": "{{user `consul_address`}}:{{user `consul_port`}}"
  }
]
```


[packer-pp]: <https://github.com/bhourigan/packer-post-processor-consul>
[bhourigan]: <https://github.com/bhourigan>
[dobuilder]: <https://www.packer.io/docs/builders/digitalocean.html>
[plugins]: <https://www.packer.io/docs/extend/plugins.html>
