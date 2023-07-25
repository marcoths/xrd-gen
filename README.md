# xrd-gen

```shell
xrd-gen is a command-line interface (CLI) to generate Crossplane Composite Resource Definitions (XRDs) from
Go structs.

Usage:
  xrd-gen [flags]

Examples:
 # Generate XRDs for all structs in the examples/deploy directory.
xrd-gen --path=examples/deploy

Flags:
  -h, --help          help for xrd-gen
  -p, --path string   path to the source CRDs
```

## Installation

### Building latest version from source

If you have a Go environment configured, you can install the latest version of `xrd-gen` from the command line.


### Building from Docker

If you have Docker installed, you can build a Docker image and run `xgen` from a container.

```shell
docker build -t xgen .
```

Then to run it as a container:

```shell
docker run --rm -it -v $(pwd):/work xgen --path=/work/examples/deploy
```