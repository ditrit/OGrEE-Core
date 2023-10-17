package models

import (
	"context"
	"encoding/base64"
	"fmt"
	"p3/repository"
	u "p3/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

const imageMaxSizeInMegaBytes = 16
const imageMaxSizeInBytes = imageMaxSizeInMegaBytes * 1024 * 1024 // 16MB, maximum document size in mongodb

// Transforms base64 encoded image into bytes and saves it into the database
func createImageFromBase64(ctx context.Context, imageBase64 string) (any, *u.Error) {
	imageSizeInBytes := base64.StdEncoding.DecodedLen(len(imageBase64))
	if imageSizeInBytes > imageMaxSizeInBytes {
		return nil, &u.Error{
			Type:    u.ErrBadFormat,
			Message: fmt.Sprintf("Image size cannot be larger than %vMB", imageMaxSizeInMegaBytes),
		}
	}

	decodedImage := make([]byte, imageSizeInBytes)
	_, err := base64.StdEncoding.Decode(decodedImage, []byte(imageBase64))
	if err != nil {
		return nil, &u.Error{Type: u.ErrBadFormat, Message: err.Error()}
	}

	imageID, uErr := repository.CreateImage(ctx, decodedImage)
	if uErr != nil {
		return nil, uErr
	}

	return imageID, nil
}

// Returns image with "id" from database
func GetImage(id string) ([]byte, *u.Error) {
	return WithTransaction(func(ctx mongo.SessionContext) ([]byte, error) {
		return repository.GetImage(ctx, id)
	})
}
