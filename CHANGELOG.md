# 0.14.0

New functionality:

* Support "monitoring packs"

  "monitoring packs" are simple tar.gz files that contain a file
  `config.tpl` + additional supporting files. When a user indicates via
  parameter that they want to use a monitoring pack, the bpm plugin goes
  through these steps:

  1. Unpacks the file into `~/.bpm/nodes/<node-id>/monitoring`
  2. Appends the file `~/.bpm/nodes/<node-id>/monitoring/config.tpl` to
     the internal filebeat configuration template
  3. Renders the combined template
  4. Starts and stops the filebeat container togeth with the other
     containers when needed

  This allows bpm to serve monitoring data and logs to all possible
  filebeat outputs. Optionally the monitoring pack can contain additional
  files like TLS certificates for authentication against monitoring
  endpoints.

  To separate the configuration of the environment (e.g. monitoring) and
  the configuration of the actual node I've added a new lifecycle step:

  - set-up-environment
  - and for completeness sake (even if currently unused): tear-down-environment

  The `set-up-environment` lifecycle step comes before `create-identity` and
  `create-configurations` and is supposed to set up the runtime environment
  for the node but not the node itself.

* Use Go templates in docker mount definitions

  This can be used to access parameters (declaratively) to determine what
  to mount.

  The first use-case for this is to use the `data-dir` parameter to be
  able to specify where blockchain data is stored:

      {
         Type: "bind",
         From: "{{ index .Node.StrParameters \"data-dir\" }}",
         To:   "/data",
      }

  In the future we can use the same technique to parameterize even more.

  BREAKING CHANGE: changes the interface of docker.go. For most plugins

Build pipeline:

* Update .gitlab-ci.yml for auto deploy of swagger file to docs

# 0.13.0

* Changed project location to `go.blockdaemon.com/bpm/sdk` which redirects to the actual repository
* Handle error when deleting node data

# 0.12.0 

New functionality:

* Docker plugins have by default a new parameter `--data-dir` that can be used to specify where to put the node data

# 0.11.0

New functionality:

* Add node.Remove method for removing nodes (from the cli)

Bug fixes:

* Stopped & started containers are now still connected to the correct docker network  

# 0.10.0

New functionality:

* Support for new plugin protocol version 1.1.0:
	* Rename the `create-secrets` call to `create-identity`. This makes the intend more clear.
	* Add a `remove-identity` call similar to `remove-config`
	* Add a `validate-parameters` call (parameters used to be validated implicitely when creating the configurations)
	* Adds plugin name to the plugin meta information
  * Make the `upgrade`, `create-identity`, `remove-identity` calls optional (in addition to `test` which has always been optional)

* New struct `DockerPlugin` that serves as a good starting point for docker based plugins

* Add the ability to launch transient containers. This is necessary for some protocols where a container needs to get
  launched temporarily during configuration, setup or upgrade.

* Split Plugin interface into multiple smaller interfaces (see details below)

  This change allows plugins to be composed of different parts that each supply
  some part of the plugin functionality.

  For example a plugin can use the existing DockerLifecycleHandler to
  manage containers but supply it's own Tester or Upgrader functionality,
  essentially mixing and matching pre-defined and custom functionality.

* Create logs directory when node is started using DockerLifecycleHandler

* Code that handles files now uses the node directory instead of the configs directory as root. Plugin developers need
  to specify the full path from root (e.g. `configs/a_config_file.json`, `logs/a_log_file.log`) instead of just the
  filename. This allows more flexibility for the plugins as they can now choose where to store files.

Bug fixes:

* For docker based packages, always expose published ports. Previously if a docker image didn't expose a port it was
  unavailable even when explicitely published.

* Removed outdated documentation and instead linked to the proper up-to-date documentation.

* When removing configuration files, also remove the `configs` directory
