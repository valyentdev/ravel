package image

import (
	"context"
	"log/slog"
	"time"
)

const MAX_TIME_SINCE_LAST_USED = 10 * time.Minute

func (i *ImagesService) StartGarbageCollection() {
	unusedSince := map[string]time.Time{}

	ticker := time.NewTicker(10 * time.Second)
	for range ticker.C {
		containers, err := i.client.ContainerService().List(context.Background())
		usedImages := map[string]struct{}{}
		for _, container := range containers {
			if err != nil {
				slog.Error("failed to get image for container %q: %v", container.ID, err)
				continue
			}
			usedImages[container.Image] = struct{}{}
		}

		images, err := i.client.ListImages(context.Background())
		if err != nil {
			slog.Error("failed to list images: %v", err)
			continue
		}
		for _, image := range images {
			if _, ok := usedImages[image.Name()]; ok {
				delete(unusedSince, image.Name())
				continue
			}

			if _, ok := unusedSince[image.Name()]; !ok {
				unusedSince[image.Name()] = time.Now()
				continue
			}

			if time.Since(unusedSince[image.Name()]) > MAX_TIME_SINCE_LAST_USED {
				if err := i.client.ImageService().Delete(context.Background(), image.Name()); err != nil {
					slog.Error("failed to remove image %q: %v", image.Name(), err)
				}
				delete(unusedSince, image.Name())
			}

		}

	}

}
