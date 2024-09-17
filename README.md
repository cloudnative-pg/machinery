[![CloudNativePG](./logo/cloudnativepg.png)](https://cloudnative-pg.io/)

# CloudNativePG - Machinery

The [CloudNativePG codebase](https://github.com/cloudnative-pg/cloudnative-pg)
is not intended to be used as a library.

This repository hosts a Go module that includes features that were moved out
of the [CloudNativePG](https://cloudnative-pg.io) codebase to make them
accessible for use in other projects.

## How to use it

As a Go module:

```
go get github.com/cloudnative-pg/machinery
```

The documentation is available on the
[pkg.go.dev](https://pkg.go.dev/github.com/cloudnative-pg/machinery) website.

## How to contribute

If you want to contribute to the source code of CloudNativePG, this is the right
place.

Have a look at the [contributing
guidelines](https://github.com/cloudnative-pg/cloudnative-pg/blob/main/contribute/README.md)
and feel free to ask in the ["dev"
chat](https://cloudnativepg.slack.com/archives/C03D68KGG65) if you have
questions or are seeking guidance.

### Testing your changes

As a prerequisite, you need to install the [Dagger
CLI](https://docs.dagger.io/quickstart/cli).

You can then check your contribution with:

```
./hack/ci.sh
```

## Users

This library is used by CloudNativePG itself and by the supporting
[`barman-cloud`](https://github.com/cloudnative-pg/barman-cloud) library.

## License

This code is released under the permissive [Apache License 2.0](./LICENSE) license.
