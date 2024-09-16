[![CloudNativePG](./logo/cloudnativepg.png)](https://cloudnative-pg.io/)

# CloudNativePG - Machinery

The [CloudNativePG code base](https://github.com/cloudnative-pg/cloudnative-pg)
is not designed to be used as a library.

This repository contains a Go module including the features that were move out
from the [CloudNativePG](https://cloudnative-pg.io) code base because they are
useful by other projects.

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