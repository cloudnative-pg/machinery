[![CloudNativePG](./logo/cloudnativepg.png)](https://cloudnative-pg.io/)

# CloudNativePG - Machinery

The [CloudNativePG code base](https://github.com/cloudnative-pg/cloudnative-pg)
is not designed to be used as a library.

This repository contains a Go module including the features that were move out
from the [CloudNativePG](https://cloudnative-pg.io) code base because they are
useful by other projects.

## Content

This library includes:

- `pkg/api` a set of reusable structures to reference a Kubernetes resource or a
  key inside a Secret or ConfigMap.

- `pkg/execlog` a set of utility functions allowing the caller to run an
  external process while dumping its output into a JSON-formatted log stream.

- `pkg/fileutils` utility functions to work with a file system.

- `pkg/log` the CloudNativePG logging infrastructure.


## How to use it

As a Go module:

```
go get github.com/cloudnative-pg/machinery
```

## Users

This library is used by CloudNativePG itself and by the supporting
[`barman-cloud`](https://github.com/cloudnative-pg/barman-cloud) library.

## License

This code is released under the permissive [Apache License 2.0](./LICENSE) license.