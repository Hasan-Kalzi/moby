package image

import (
	"testing"

	"github.com/moby/moby/client"
	"github.com/moby/moby/v2/integration/internal/container"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"
	"gotest.tools/v3/poll"
)

func TestCommitChangeLabels(t *testing.T) {
	// Register cleanup for the containers and images created by this test.
	ctx := setupTest(t)
	apiClient := testEnv.APIClient()

	// Start a container with the original label.
	containerID := container.Run(ctx, t, apiClient,
		container.WithCmd("true"),
		func(c *container.TestContainerConfig) {
			c.Config.Labels = map[string]string{"some": "label"}
		},
	)

	// Wait for the command to finish successfully before committing.
	poll.WaitOn(t, container.IsSuccessful(ctx, apiClient, containerID))

	// Override the label in the image created by the commit.
	image, err := apiClient.ContainerCommit(ctx, containerID,
		client.ContainerCommitOptions{
			Changes: []string{"LABEL some=label2"},
		},
	)
	assert.NilError(t, err)

	// The new image should contain the updated label.
	inspect, err := apiClient.ImageInspect(ctx, image.ID)
	assert.NilError(t, err)
	assert.Check(t, is.DeepEqual(
		inspect.Config.Labels,
		map[string]string{"some": "label2"},
	))

	// Committing must not change the source container's label.
	source := container.Inspect(ctx, t, apiClient, containerID)
	assert.Check(t, is.DeepEqual(
		source.Config.Labels,
		map[string]string{"some": "label"},
	))
}