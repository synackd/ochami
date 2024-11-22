OCHAMI-CONFIG(5) "OpenCHAMI" "ochami: The OpenCHAMI CLI Tool"

# NAME

config.yaml - ochami CLI configuration file

# DESCRIPTION

*ochami* supports different config file formats including _yaml_, _json_, and
_toml_, but YAML is the default. Configuration options can be set via the
*ochami config* command.

# CONFIGURATION

## Global Options

These configuration options are global configuration options.

*default-cluster:* _cluster_name_
	The name of the default cluster to use when *--cluster* is not specified on
	the command line. A cluster configuration must exist for _cluster_name_ or
	further commands will fail.

*log*
	Logging options.

	*format:* _format_
		The format of log messages.

		Default: *json*
		Supported:
		- _basic_
		- _json_
		- _rfc3339_

	*level:* _level_
		Logging level.

		Default: *warning*
		Supported:
		- _info_
		- _warning_
		- _debug_

## Cluster Configuration

These configuration options apply only to cluster configuration, i.e. under the
*clusters* key. The value for the *cluster* key is an array with each item in
the array containing the below configuration options.

*cluster*
	The key containing cluster configuration subkeys.

	*base-uri:* _base_uri_
		The base URI for the OpenCHAMI services for the cluster.

*name:* _cluster_name_
	The name of the cluster. This is what *--cluster* and the *default-cluster*
	key use to identify the cluster.

# EXAMPLE

```
clusters:
    - cluster:
        base-uri: https://foobar.openchami.cluster
      name: foobar
default-cluster: foobar
log:
    format: json
    level: debug
```

# FILES

_~/.config/ochami/config.yaml_

# AUTHOR

Written by Devon T. Bautista and maintained by the OpenCHAMI developers.

# SEE ALSO

*ochami-config*(1)

; Vim modeline settings
; vim: set tw=80 noet sts=4 ts=4 sw=4 syntax=scdoc:
