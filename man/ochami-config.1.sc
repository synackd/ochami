OCHAMI-CONFIG(1) "OpenCHAMI" "Manual Page for ochami-config"

# NAME

ochami-config - Manage configuration for ochami CLI

# SYNOPSIS

ochami config [OPTIONS] COMMAND

# COMMANDS

## cluster

Manage cluster configurations.

Subcommands for this command are as follows:

*delete* _cluster_name_
	Delete _cluster_name_ configuration from config file.

*set* [--base-uri _base_uri_] [--default] _cluster_name_
	Add or set configuration for a cluster.

	This command accepts the following options:

	*-u, --base-uri* _base_uri_
		Specify the base URI of OpenCHAMI services for the cluster.

		*ochami* will use this to concatenate endpoint information to when
		communicating with this cluster's OpenCHAMI services.

	*-d, --default*
		Set this cluster as the default cluster. This means that if *--cluster*
		is not specified on the command line, this cluster's configuration is
		used.

## show

Show the current configuration. This command can be used to generate a
configuration file populated with the default values.

This command accepts the following options:

*-f, --format* _format_
	Format of config output.

	Default: *json*
	Supported:
	- _json_
	- _yaml_

# AUTHOR

Written by Devon T. Bautista and maintained by the OpenCHAMI developers.

# SEE ALSO

*ochami-config*(5)

; Vim modeline settings
; vim: set tw=80 noet sts=4 ts=4 sw=4 syntax=scdoc:
