# 0.11.0

New functionality:

* `status` call in docker based plugins now checks if the docker network exists. Previously it just crashed. 

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
