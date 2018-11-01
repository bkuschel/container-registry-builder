package icrbuild

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/IBM-Cloud/bluemix-go/api/container/registryv1"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

// Use the standard Docker APIs to leverage standard CLI impementation
type Builder struct {
	client.APIClient
	registryClient *IBMRegistrySession
}

type builderCLI struct {
	command.DockerCli
	builder *Builder
}

// NewBuildClient with the IBM Cloud Container Registry CLIs
func NewBuilder(registryClient *IBMRegistrySession) *Builder {
	return &Builder{
		registryClient: registryClient,
	}
}

// ImageBuild satisfies the Docker Client interface for performing an image build
func (o *Builder) ImageBuild(_ context.Context, buildctx io.Reader, opts types.ImageBuildOptions) (types.ImageBuildResponse, error) {
	var (
		imageBuildRequest registryv1.ImageBuildRequest
		buildArgBytes     []byte
		buildResponse     types.ImageBuildResponse
		pr                *io.PipeReader
		pw                *io.PipeWriter
		tag               string
		err               error
	)

	if len(opts.Tags) >= 1 {
		tag = opts.Tags[0]
	}

	if opts.BuildArgs != nil && len(opts.BuildArgs) > 0 {
		buildArgBytes, err = json.Marshal(opts.BuildArgs)
		if err != nil {
			return buildResponse,errors.Wrap(err, "Unable to marshal build args as json")
		}
	}

	imageBuildRequest = registryv1.ImageBuildRequest{
		T:          tag,
		Dockerfile: opts.Dockerfile,
		Buildargs:  fmt.Sprintf("%s", buildArgBytes),
		Pull:       opts.PullParent,
		Nocache:    opts.NoCache,
	}

	pr, pw = io.Pipe()
	go func() {
		if err := o.registryClient.Builds.ImageBuild(imageBuildRequest, buildctx, o.registryClient.BuildTargetHeader, pw); err != nil {
			pw.Write([]byte(fmt.Sprintf(`{"errorDetail":{"message":"%v"}}`, err)))
		}
		pw.Close()
	}()

	buildResponse.Body = pr

	return buildResponse, nil

}

// DaemonHost stub to Satisfy APIClient API (unused)
func (o *Builder) DaemonHost() string {
	return ""
}

func (b *builderCLI) Client() client.APIClient {
	return b.builder
}

func (b *builderCLI) ConfigFile() *configfile.ConfigFile {
	return &configfile.ConfigFile{}
}