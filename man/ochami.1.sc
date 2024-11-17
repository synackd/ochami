OCHAMI(1) "OpenCHAMI" "Manual Page for ochami"

# NAME

ochami - OpenCHAMI command line interface

# SYNOPSIS

ochami [OPTIONS] COMMAND

# DESCRIPTION

*ochami* is a command line interface tool that interacts with the OpenCHAMI API.
It is meant to provide a convenience so one need not craft *curl* incantations,
but also provide additional features.

List of available commands:

[[ *Command*
:< *Description*
|  *bss*
:  Communicate with the Boot Script Service (BSS)
|  *cloud-init*
:  Manage cloud-init configurations
|  *discover*
:  Simulate discovery of BMCs and nodes to populate SMD by reading an input file
|  *smd*
:  Communicate with the State Management Database (SMD)
|  *config*
:  Manage ochami CLI configuration, including cluster configuration

## Top-Level Commands

Top-level commands can be thought of as more abstract and admin-friendly.
These commands tend not to correspond 1:1 with single HTTP requests. For
instance, the *discover* command makes multiple requests to different SMD
endpoints, some of them iteratively.

## Low-Level Commands

Commands that correspond to single HTTP requests ("low-level" commands) are
organized under metacommands corresponding to the service that they send the
request to, under subcommands corresponding to the endpoint of the request. For
instance, the *smd* command has the *component* subcommand that deals with SMD's
_/Components_ endpoint. This subcommand has further subcommands *add*, *delete*,
and *get* which send POST, DELETE, and GET HTTP requests, respectively. These
commands are primarily used for manually adding/getting/modifying/deleting data
structures, e.g. when troubleshooting.

# GETTING STARTED

Upon first running *ochami*, a config file will need to be generated and a basic
cluster configuration will need to be specified. Both can be done with the
following command:

```
ochami config cluster set -d -u https://foobar.openchami.cluster foobar
```

This will create a cluster called _foobar_ and set its base URI to
_https://foobar.openchami.cluster_, placing this config in
_~/.config/ochami/config.yaml_. Since *ochami* supports multiple cluster
configurations, the _-d_ tells *ochami* to set this cluster as the default
cluster, which means that this cluster's configuration will be used if
_--cluster_ is not specified on the command line.

Once the cluster configuration has been specified, one will need to store a
token to be able to be used to authenticate to protected endpoints without
having to specify _--token_ every time. *ochami* will look for an environment
variable named *\<CLUSTER_NAME\>_ACCESS_TOKEN*, where *\<CLUSTER_NAME\>* is
replaced by the name of the cluster in all capital letters and dashes replaced
by underscores. In the above example, this variable would be named
*FOOBAR_ACCESS_TOKEN*:

```
export FOOBAR_ACCESS_TOKEN=...
```

Once these steps are completed, *ochami* should be ready to use with cluster
_foobar_.

# GLOBAL OPTIONS

*-u, --base-uri* _uri_
	Specify the base URI to use when contacting OpenCHAMI services. Overrides
	the base URI specified in a config file.

*--cacert* _cacert_
	Specify the path to a certificate authority (CA) certificate file to use to
	verify TLS certificates. Must be PEM-formatted.

*--cluster* _cluster_name_
	Specify the name of a cluster to use. The cluster corresponding to the
	passed cluster name must exist in the config file.

*-c, --config* _config_file_
	Specify the path to a config file to use. By default, this is
	_~/.config/ochami/config.yaml_. The format of this file is assumed to be
	YAML unless either the file extension differs from _yaml_ (in which case
	*ochami* attempts to infer the format from the file extention) or
	_--payload-format_ is specified. See the description of _--payload-format_
	for supported config file formats.

*--config-format* _format_
	Explicitly specify the format of the default config file or the config file
	passed with _--config_. Supported config formats are:

	- _json_
	- _yaml_

*--ignore-config*
	Do not read configuration from any configuration file.

*-k, --insecure*
	Do not verify TLS certificates.

*--log-format* _format_
	Specify the format of log messages, overriding what is set in the config
	file. Defaults to _json_.

	Supported log formats are:

	- _basic_
	- _json_
	- _rfc3339_

*-l, --log-level* _level_
	Specify the level to print log messages at, overriding what is set in the
	config file. Defaults to _warning_.

	Supported log levels are:

	- _info_
	- _warning_
	- _debug_

*-t, --token* _token_
	Access token to include in request headers for authenticated to protected
	service endpoints. Overrides token set in environment variable.

# FILES

_~/.config/ochami/config.yaml_
	The ochami CLI configuration file. Generated if non-existent upon
	command invocation.

# AUTHOR

Written by Devon T. Bautista and maintained by the OpenCHAMI developers.

# SEE ALSO

*ochami-bss*(1), *ochami-discover*(1), *ochami-smd*(1)

; Vim modeline settings
; vim: set tw=80 noet sts=4 ts=4 sw=4 syntax=scdoc:
