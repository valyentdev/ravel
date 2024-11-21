package images

import (
	"github.com/spf13/cobra"
	"github.com/valyentdev/ravel/cmd/ravel/util"
	"github.com/valyentdev/ravel/runtime"
)

func NewImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Manage images",
	}

	cmd.AddCommand(newListImagesCmd())
	cmd.AddCommand(newPullImageCmd())
	cmd.AddCommand(newDeleteImageCmd())

	return cmd
}

func newListImagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List images",
		RunE:  listImages,
	}

	return cmd
}

func listImages(cmd *cobra.Command, args []string) error {
	client := util.GetAgentClient(cmd)

	images, err := client.ListImages(cmd.Context())
	if err != nil {
		return err
	}

	cmd.Println("NAME")
	for _, image := range images {
		cmd.Println(image.Name)
	}

	return nil
}

func newPullImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pull <ref>",
		Short: "Pull an image",
		RunE:  pullImage,
	}

	return cmd
}

func pullImage(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	ref := args[0]

	client := util.GetAgentClient(cmd)

	cmd.Println("Pulling image...")
	_, err := client.PullImage(cmd.Context(), runtime.PullImageOptions{
		Ref: ref,
	})
	if err != nil {
		return err
	}

	cmd.Println("Image pulled")
	return nil
}

func newDeleteImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <ref>",
		Short: "Delete an image",
		RunE:  deleteImage,
	}

	return cmd
}

func deleteImage(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	ref := args[0]

	client := util.GetAgentClient(cmd)

	cmd.Println("Deleting image...")
	err := client.DeleteImage(cmd.Context(), ref)
	if err != nil {
		return err
	}

	cmd.Println("Image deleted")
	return nil
}
