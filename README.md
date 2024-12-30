# ochami: OpenCHAMI Command Line Interface

<!-- Text width is 80, only use spaces and use 4 spaces instead of tabs -->
<!-- vim: set et sta tw=80 ts=4 sw=4 sts=0: -->

`ochami` is the command line interface to interact with the API of OpenCHAMI
services, especially the [State Management Database
(SMD)](https://github.com/OpenCHAMI/smd) and the [Boot Script Service
(BSS)](https://github.com/OpenCHAMI/bss). The tool is meant to ease interaction
with the API so one need not be proficient in `curl`.

## Getting Started

See [**Building**](#building) for instructions on how to build `ochami`. Then,
continue with how to use it.

### 1. Generating a Configuration File

By default, `ochami` reads the config file from
`~/.config/ochami/config.yaml`[^config-format][^config-file]. If it does not
exist, the user will be asked to create it.

[^config-format]: `ochami` supports all config file formats that
    [Viper](https://github.com/spf13/viper) supports. Unless `--config-format`
    is passed, `ochami` tries to determine the format via the file extension. By
    default, the YAML format is used.
[^config-file]: `-c` or `--config` can be used to change the config file path.

Run the following command to generate the config file and show the default
configuration:

```bash
$ ochami config show
Config file /home/user/.config/ochami/config.yaml does not exist. Create it? [yN]: y
log:
    format: json
    level: warning

```

### 2. Adding a Cluster

Next, `ochami` needs to be told how to contact the Ochami services. The
configuration file could be edited to do this, but `ochami` provides the
`config` command to edit configuration.

Run the following command to add a default cluster configuration called `foobar`
whose base URI is `https://foobar.openchami.cluster`:

```bash
ochami config cluster set foobar --default --base-uri https://foobar.openchami.cluster
```

**NOTE:** Since `ochami` supports multiple cluster configurations, `--default`
makes this cluster the default, meaning if `--cluster` is not specified on the
command line, this cluster's configuration will be used.

Now, when the configuration is shown, we should see the new cluster's details:

```bash
$ ochami config show
clusters:
    - cluster:
        base-uri: https://foobar.openchami.cluster
      name: foobar
default-cluster: foobar
log:
    format: json
    level: warning

```

### 3. Testing Unauthenticated Cluster Access

Test access by contacting an API endpoint not requiring an access token:

```bash
$ ochami bss status
{"bss-status":"running"}

```

### 4. Setting Access Token for Cluster

Since `ochami` supports multiple cluster configurations, it supports reading
environment variables corresponding to the cluster for the access token. This
can be overridden by using `--token`. Since our cluster is named "foobar", we
need to set `FOOBAR_ACCESS_TOKEN` to the value of the token so `ochami` can read
it when communicating with this cluster.

```bash
export FOOBAR_ACCESS_TOKEN=eyJhbGc...
```

Note that the format of the environment variable that `ochami` reads for the
access token is `<CLUSTER_NAME>_ACCESS_TOKEN` where `<CLUSTER_NAME>` is the
value of the cluster name (`name` in the config file specified with `--cluster`,
or `default-cluster` in the config file, the former taking precedence over the
latter) in all capitals and with dashes (-) and spaces substituted with
underscores (_).

### 5. Testing Authenticated Cluster Access

Now, we should be able to contact the API on an endpoint that requires
authentication:

```bash
$ ochami bss boot params get
null

```

## Building

### Goreleaser

Goreleaser is the way ochami gets built for releases, and is the officially
supported build method for troubleshooting.

```bash
export GOVERSION=$(go env GOVERSION)
export BUILD_HOST=$(hostname)
export BUILD_USER=$(whoami)
goreleaser build --clean --snapshot --single-target
```

Remove `--single-target` to build for all targets.

### Make

Make provides convenient and quick building for fast iteration and development.

Linker flags are used to embed build metadata into the binary. Building can
simply be done via:

```bash
make
```

On Unix-like systems, one can also install the binary, man pages, and
completions:

```bash
sudo make install
```

## Container

### Pulling

```bash
docker pull ghcr.io/synackd/ochami:latest
```

### Building

There are two dockerfiles for two different purposes.

- **Dockerfile** is for manual building and is intended for building locally. It
  uses a multi-stage build, the first stage building from source and the second
  stage copying the binary from the first stage.
- **goreleaser.dockerfile** is used by Goreleaser for CI. It assumes the binary
  has already been built and copies it into the container.

To build the ochami container (with `dirty` tag):

```bash
docker build . --tag ochami:dirty
```

### Running

```bash
docker run ghcr.io/synackd/ochami:latest ochami --ignore-config help
```
The above incantation will print out the command line's help message.
