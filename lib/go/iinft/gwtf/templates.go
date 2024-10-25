package gwtf

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
)

func splitByWidthMake(str string, size int) []string {
	strLength := len(str)
	splitedLength := int(math.Ceil(float64(strLength) / float64(size)))
	splited := make([]string, splitedLength)
	var start, stop int
	for i := 0; i < splitedLength; i++ {
		start = i * size
		stop = start + size
		if stop > strLength {
			stop = strLength
		}
		splited[i] = str[start:stop]
	}
	return splited
}

func fileAsImageData(path string) (string, error) {
	f, _ := os.Open(path)

	defer f.Close()

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read imageFile %s, %w", path, err)
	}

	return contentAsImageDataURL(content), nil
}

func contentAsImageDataURL(content []byte) string {
	contentType := http.DetectContentType(content)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)

	return "data:" + contentType + ";base64, " + encoded
}

func fileAsBase64(path string) (string, error) {
	f, _ := os.Open(path)

	defer f.Close()

	// Read entire JPG into byte slice.
	reader := bufio.NewReader(f)
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("could not read file %s, %w", path, err)
	}

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)

	return encoded, nil
}

// UploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (f *GoWithTheFlow) UploadFile(ctx context.Context, filename, accountName string) error {
	content, err := fileAsBase64(filename)
	if err != nil {
		return err
	}

	return f.UploadString(ctx, content, accountName)
}

func getURL(url string) ([]byte, error) {
	resp, err := http.Get(url) //nolint:gosec // inherited from GWTF
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// DownloadAndUploadFile reads a file, base64 encodes it and chunk upload to /storage/upload
func (f *GoWithTheFlow) DownloadAndUploadFile(ctx context.Context, url, accountName string) error {
	body, err := getURL(url)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(body)
	return f.UploadString(ctx, encoded, accountName)
}

// DownloadImageAndUploadAsDataURL download an image and upload as data url
func (f *GoWithTheFlow) DownloadImageAndUploadAsDataURL(ctx context.Context, url, accountName string) error {
	body, err := getURL(url)
	if err != nil {
		return err
	}
	content := contentAsImageDataURL(body)

	return f.UploadString(ctx, content, accountName)
}

// UploadImageAsDataURL will upload a image file from the filesystem into /storage/upload of the given account
func (f *GoWithTheFlow) UploadImageAsDataURL(ctx context.Context, filename, accountName string) error {
	content, err := fileAsImageData(filename)
	if err != nil {
		return err
	}

	return f.UploadString(ctx, content, accountName)
}

// UploadString will upload the given string data in 1mb chunks to /storage/upload of the given account
func (f *GoWithTheFlow) UploadString(ctx context.Context, content, accountName string) error {
	// unload previous content if any.
	if _, err := f.Transaction(`
	transaction {
		prepare(signer: AuthAccount) {
			let path = /storage/upload
			let existing = signer.load<String>(from: path) ?? ""
			log(existing)
		}
	}
	  `).SignProposeAndPayAs(accountName).RunE(ctx); err != nil {
		return err
	}

	parts := splitByWidthMake(content, 1_000_000)
	for _, part := range parts {
		if _, err := f.Transaction(`
		transaction(part: String) {
			prepare(signer: AuthAccount) {
				let path = /storage/upload
				let existing = signer.load<String>(from: path) ?? ""
				signer.save(existing.concat(part), to: path)
				log(signer.address.toString())
				log(part)
			}
		}
			`).SignProposeAndPayAs(accountName).StringArgument(part).RunE(ctx); err != nil {
			return err
		}
	}

	return nil
}
