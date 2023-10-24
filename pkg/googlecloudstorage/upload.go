package googlecloudstorage

import (
	"context"
	"encoding/base64"
	"io"

	"cloud.google.com/go/storage"
)

func uploadToGCS(client *storage.Client, bucketName, objectName, data string) error {
	wc := client.Bucket(bucketName).Object(objectName).NewWriter(context.Background())
	b, _ := base64.StdEncoding.DecodeString(data)
	if _, err := io.WriteString(wc, string(b)); err != nil {
		return err
	}
	return wc.Close()
}
