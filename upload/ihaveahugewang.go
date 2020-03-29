// Package upload contains functions to upload files on variuos file hostings.
package upload

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	ihaveahugewangUploader "github.com/astravexton/ihaveahuge.wang.cmd"
	"github.com/astravexton/irchuu/config"
	"github.com/astravexton/irchuu/paths"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

// Ihaveahugewang uploads a Telegram media file to ihaveahuge.wang
func Ihaveahugewang(bot *tgbotapi.BotAPI, id string, c *config.Telegram) (url string, err error) {
	file, err := bot.GetFileDirectURL(id)
	if err != nil {
		return
	}
	fileStrings := strings.Split(file, "/")
	fileName := strings.Split(fileStrings[len(fileStrings)-1], ".")

	var ext string
	if len(fileName) > 1 {
		ext = "." + fileName[len(fileName)-1]
	}
	localUrl := path.Join(c.DataDir, id+ext)
	// if it is already downloaded, just upload the local copy
	if paths.Exists(localUrl) {
		return uploadLocalFileWang(localUrl, c)
	}
	return uploadRemoteFileWang(file, localUrl, id, fileStrings[len(fileStrings)-1], c)
}

// uploadLocalFileWang actually uploads the file to a pomf clone using HTTP POST with
// multipart/form-data mime. It also reads the whole file to memory because of
// the current implementation of Go's multipart.
func uploadLocalFileWang(file string, c *config.Telegram) (url string, err error) {
	ct, bd := ihaveahugewangUploader.PrepareUploadBody(file)
	data, err := ihaveahugewangUploader.UploadToSite(ct, bd)
	url = fmt.Sprintf("https://ihaveahuge.wang/i/%s", data.Slug)
	return
}

// uploadRemoteFileWang downloads a file from Telegram and uploads it to a
// pomf clone using HTTP POST with multipart/form-data mime. It also reads the
// whole file to memory because of the current implementation of Go's multipart.
func uploadRemoteFileWang(file string, localUrl string, id string, name string, c *config.Telegram) (url string, err error) {
	downloadable, err := http.Get(file)
	if err != nil {
		return
	}
	defer downloadable.Body.Close()

	res, err := os.Create(localUrl)
	if err != nil {
		return "", err
	}

	_, err = io.Copy(res, downloadable.Body)
	res.Close()
	if err != nil {
		return "", err
	}

	ct, bd := ihaveahugewangUploader.PrepareUploadBody(localUrl)
	data, err := ihaveahugewangUploader.UploadToSite(ct, bd)
	url = fmt.Sprintf("https://ihaveahuge.wang/i/%s", data.Slug)

	return
}

// {"filetype":"image/png","slug":"fdee78ed"}
// ihaveahugewangResult is the data ihaveahugewangResult returns in JSON.
type ihaveahugewangResult struct {
	FileType string
	Slug     string
}
