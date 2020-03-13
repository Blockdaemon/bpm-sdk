# 0.10.0

New functionality:

* Support for new plugin protocol version 1.1.0:
	* Rename the `create-secrets` call to `create-identity`. This makes the intend more clear.
	* Add a `remove-identity` call similar to `remove-config`
	* Add a `validate-parameters` call (parameters used to be validated implicitely when creating the configurations)
	* Adds plugin name to the plugin meta information

* Add the ability to launch transient containers. This is necessary for some protocols where a container needs to get launched temporarily during configuration, setup or upgrade.

* Split Plugin interface into multiple smaller interfaces (see details below)

  This change allows plugins to be composed of different parts that each supply
  some part of the plugin functionality.

  For example a plugin can use the existing DockerLifecycleHandler to
  manage containers but supply it's own Tester or Upgrader functionality,
  essentially mixing and matching pre-defined and custom functionality.

  Previously this was only possible by "mimicking" inheritance like this
  which is hard to understand, not very flexible and not go-idiomatic:

  	  // PolkadoDockerPlugin uses DockerPlugin but overwrites functions to add custom test functionality
  	  type PolkadotDockerPlugin struct {
  	  	plugin.Plugin
  	  }

  	  // Test the node
  	  func (d PolkadotDockerPlugin) Test(currentNode node.Node) (bool, error) {
  	  	if err := runAllTests(); err != nil {
  	  		return false, err
  	  	}
  	  	return true, nil
  	  }


* Create logs directory when node is started using DockerLifecycleHandler

Bug fixes:

* For docker based packages, always expose published ports. Previously if a docker image didn't expose a port it was unavailable even when explicitely published.

* Removed outdated documentation and instead linked to the proper up-to-date documentation.


