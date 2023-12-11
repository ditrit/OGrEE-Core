package models

import (
	"context"
	"fmt"
	"p3/repository"
	u "p3/utils"

	"github.com/elliotchance/pie/v2"
	"github.com/vincent-petithory/dataurl"
)

const imageMaxSizeInMegaBytes = 16
const imageMaxSizeInBytes = imageMaxSizeInMegaBytes * 1024 * 1024 // 16MB, maximum document size in mongodb
const imageType = "image"

var acceptedImageFormats = []string{imageType + "/png", imageType + "/jpeg", imageType + "/webp"}

// Transforms data uri with a base64 encoded image into bytes and saves it into the database
func createImageFromDataURI(ctx context.Context, dataURI string) (any, *u.Error) {
	dataURL, err := dataurl.DecodeString(dataURI)
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: "Incorrect data uri on image"}
	}

	if dataURL.Type != imageType || !pie.Contains(acceptedImageFormats, dataURL.ContentType()) {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: "Image format not supported"}
	}

	imageSizeInBytes := len(dataURL.Data)
	if imageSizeInBytes > imageMaxSizeInBytes {
		return nil, &u.Error{
			Type:    u.ErrBadFormat,
			Message: fmt.Sprintf("Image size cannot be larger than %vMB", imageMaxSizeInMegaBytes),
		}
	}

	imageID, uErr := repository.CreateImage(ctx, u.Image{
		MIMEType: dataURL.ContentType(),
		Data:     dataURL.Data,
	})
	if uErr != nil {
		return nil, uErr
	}

	return imageID, nil
}

// Returns image with "id" from database
func GetImage(id string) (*u.Image, *u.Error) {
	return repository.GetImage(id)
}
