package node

import (
	"io/ioutil"
	"encoding/json"
	"path"
	homedir "github.com/mitchellh/go-homedir"
	"gitlab.com/Blockdaemon/bpm-sdk/internal/util"
)

type Node struct {
	NodeGID string `json:"node_gid"`
	BlockchainGID string `json:"blockchain_gid"`

	Environment  string `json:"environment"`
	NetworkType  string `json:"network_type"`
	NodeSubtype  string `json:"node_subtype"`
	ProtocolType string `json:"protocol_type"`

	Config map[string]interface{} `json:"config"`
	Secrets map[string]interface{} // No json here, never serialize secrets!

	CurrentVersion string

	baseDir string
}

func (c Node) DockerNetworkName() string {
	return "bd-" + c.NodeGID
}

func (c Node) ContainerName(containerName string) string {
	return "bd-" + c.NodeGID + "-" + containerName
}

func (c Node) VolumeName(volumeName string) string {
	return "bd-" + c.NodeGID + "-" + volumeName
}

func (c Node) NodeDirectory() string {
	expandedBaseDir, err := homedir.Expand(c.baseDir)
	if err != nil {
		panic(err) // Should never happen because, at this stage, the directory should already be created
	}

	return path.Join(expandedBaseDir, "nodes", c.NodeGID)
}

func (c Node) CurrentVersionFile() string {
	return path.Join(c.NodeDirectory(), "version")
}

func (c Node) NodeFile() string {
	return path.Join(c.NodeDirectory(), "node.json")
}

func (c Node) ConfigsDirectory() string {
	return path.Join(c.NodeDirectory(), "configs")
}

func (c Node) SecretsDirectory() string {
	return path.Join(c.NodeDirectory(), "secrets")
}

func (c Node) WritePluginVersion(version string) error {
	c.CurrentVersion = version

	return ioutil.WriteFile(c.CurrentVersionFile(), []byte(version), 0644)
}

func LoadNode(baseDir, nodeGID string) (Node, error) {
	var node Node

	node.NodeGID = nodeGID
	node.baseDir = baseDir

	// Create node directories if they don't exist yet
	_, err := util.MakeDirectory(node.SecretsDirectory())
	if err != nil {
		return node, err
	}

	_, err = util.MakeDirectory(node.ConfigsDirectory())
	if err != nil {
		return node, err
	}

	// Load node data
	nodeData, err := ioutil.ReadFile(node.NodeFile())
	if err != nil {
		return node, err
	}

	err = json.Unmarshal(nodeData, &node)
	if err != nil {
		return node, err
	}

	// Load secrets
	node.Secrets = make(map[string]interface{})

	files, err := ioutil.ReadDir(node.SecretsDirectory())
    if err != nil {
    	return node, err
    }

    for _, f := range files {
    	if !f.IsDir() {
	    	secret, err := ioutil.ReadFile(path.Join(node.SecretsDirectory(), f.Name()))
	    	if err != nil {
	    		return node, err
	    	}

	    	node.Secrets[f.Name()] = string(secret)
    	}
    }

    // Load installed version (if there is any)
    exists, err := util.FileExists(node.CurrentVersionFile())
    if err != nil {
    	return node, err
    }

    if exists {
    	versionData, err := ioutil.ReadFile(node.CurrentVersionFile())
    	if err != nil {
    		return node, err
    	}

    	node.CurrentVersion = string(versionData)
    }

	return node, nil
}

