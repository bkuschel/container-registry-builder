package icrbuild

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/command/image"
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type BuildFlags struct {
	NoCache   bool
	Pull      bool
	Quiet     bool
	BuildArgs []string
	File      string
	Tag       string
}

type BuildOptions struct {
	In   io.Reader
	Out  io.Writer
	Err  io.Writer
	Cmd  *cobra.Command
	Args []string

	Flags BuildFlags
}

type BuildRunner interface {
	Run(cmd *cobra.Command, args []string) error
}

// NewBuildOptions
func NewBuildOptions(in io.Reader, out io.Writer, err io.Writer) *BuildOptions {
	return &BuildOptions{
		In:  in,
		Out: out,
		Err: err,
	}
}

func (o *BuildOptions) Run(cmd *cobra.Command, args []string) error {

	var (
		registryClient                    *IBMRegistrySession
		imageName,buildContext string
		err                               error
		cli                               *builderCLI
		ccmd                              *cobra.Command
	)

	if !reference.ReferenceRegexp.MatchString(o.Flags.Tag) {
		return errors.Errorf("Image Name is not correct format!")
	}

	registryClient, imageName, err = NewRegistryClient(o.Flags.Tag)
	if err != nil {
		return errors.Wrap(err, "Unable to Connect to IBM Cloud")
	}

	logrus.Debugf("Running IBM Container Registry build: context: %s, dockerfile: %s", args[0], o.Flags.File)

	buildContext, err = filepath.Abs(o.Args[0])
	if err != nil {
		logrus.Errorf("Error parsing build context: %v", err)
		return errors.Wrap(err, "Docker build Context error! Check supplied context path")
	}

	cli = &builderCLI{*command.NewDockerCli(os.Stdin, os.Stdout, os.Stderr, false), NewBuilder(registryClient)}

	ccmd = image.NewBuildCommand(cli)

	ccmd.Flags().Set("tag", imageName)
	ccmd.Flags().Set("no-cache", strconv.FormatBool(o.Flags.NoCache))
	ccmd.Flags().Set("quiet", strconv.FormatBool(o.Flags.Quiet)) // TODO fix quiet returning the whole result at the end
	ccmd.Flags().Set("pull", strconv.FormatBool(o.Flags.Pull))
	ccmd.Flags().Set("disable-content-trust", "true")
	ccmd.Flags().Set("file", o.Flags.File)
	for _, buildFlag := range o.Flags.BuildArgs {
		ccmd.Flags().Set("build-arg", buildFlag)
	}

	err = ccmd.RunE(nil, []string{buildContext})

	return err
}
