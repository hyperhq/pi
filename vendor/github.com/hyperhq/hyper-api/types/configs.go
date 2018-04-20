package types

import (
	"github.com/hyperhq/hyper-api/types/container"
	"github.com/hyperhq/hyper-api/types/network"
)

// configs holds structs used for internal communication between the
// frontend (such as an http server) and the backend (such as the
// docker daemon).

// ContainerCreateConfig is the parameter set to ContainerCreate()
type ContainerCreateConfig struct {
	Name             string
	Config           *container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
	AdjustCPUShares  bool
}

// ContainerRmConfig holds arguments for the container remove
// operation. This struct is used to tell the backend what operations
// to perform.
type ContainerRmConfig struct {
	ForceRemove, RemoveVolume, RemoveLink bool
}

// ContainerCommitConfig contains build configs for commit operation,
// and is used when making a commit with the current state of the container.
type ContainerCommitConfig struct {
	Pause   bool
	Repo    string
	Tag     string
	Author  string
	Comment string
	// merge container config into commit config before commit
	MergeConfigs bool
	Config       *container.Config
}

// ExecConfig is a small subset of the Config struct that holds the configuration
// for the exec feature of docker.
type ExecConfig struct {
	User         string `json:"user"`  // User that will run the command
	Privileged   bool   `json:"privileged"`  // Is the container in privileged mode
	Tty          bool   `json:"tty"`  // Attach standard streams to a tty.
	AttachStdin  bool   `json:"attachStdin"`  // Attach the standard input, makes possible user interaction
	AttachStderr bool   `json:"attachStderr"`  // Attach the standard output
	AttachStdout bool   `json:"attachStdout"`  // Attach the standard error
	Detach       bool   `json:"detach"`  // Execute in detach mode
	DetachKeys   string `json:"detachKeys"`  // Escape keys for detach
	Cmd          []string `json:"cmd"` // Execution commands and args
}
