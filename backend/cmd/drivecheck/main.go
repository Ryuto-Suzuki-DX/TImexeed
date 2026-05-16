package main

import (
	"context"
	"fmt"
	"strings"

	"timexeed/backend/internal/storage"
)

func main() {
	ctx := context.Background()

	service, err := storage.NewGoogleDriveServiceFromEnv(ctx)
	if err != nil {
		panic(err)
	}

	folderID, err := service.ParseFolderID("https://drive.google.com/drive/folders/10II_cvD7lTlmX6OvcLpz8eZT3Shyp4NU")
	if err != nil {
		panic(err)
	}

	uploaded, err := service.UploadFile(
		ctx,
		folderID,
		"timexeed_drive_check.txt",
		"text/plain",
		strings.NewReader("drive upload check"),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("OK: id=%s name=%s url=%s\n", uploaded.DriveFileID, uploaded.FileName, uploaded.FileURL)
}
