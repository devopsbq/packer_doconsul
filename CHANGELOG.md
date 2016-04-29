# Changelog

### 0.2.0

* Better parsing of URL for Consul Address. Now you can pass in a URL with no port, and it will take the default 8500 Consul API port. You can also pass in the scheme (http or https), but packer-doconsul will now ignore it (instead of failing, like before); it will only pay attention to the `consul_scheme` parameter.

### 0.1.0

* Initial release.