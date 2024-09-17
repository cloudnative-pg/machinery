[![CloudNativePG](./logo/cloudnativepg.png)](https://cloudnative-pg.io/)

# CloudNativePG - Machinery

The [CloudNativePG codebase](https://github.com/cloudnative-pg/cloudnative-pg)is not intended to be used as a library.

This repository hosts a Go module that includes features that were moved out
of the [CloudNativePG](https://cloudnative-pg.io) codebase to make them
accessible for use in other projects.

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