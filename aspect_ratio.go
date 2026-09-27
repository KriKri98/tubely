package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

type ffprobeStruct struct {
	Streams []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"streams"`
}

func getVideoAspectRatio(filePath string) (string, error) {
	ffprobeCMD := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmdOut := &bytes.Buffer{}
	ffprobeCMD.Stdout = cmdOut

	err := ffprobeCMD.Run()
	if err != nil {
		return "", fmt.Errorf("unable to run command ffprobe")
	}

	videoInformation := ffprobeStruct{}
	err = json.Unmarshal(cmdOut.Bytes(), &videoInformation)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal data")
	}
	aspectRatio := ""
	height := videoInformation.Streams[0].Height
	width := videoInformation.Streams[0].Width

	if height/9 == width/16 {
		aspectRatio = "landscape"
	} else if width/9 == height/16 {
		aspectRatio = "portrait"
	} else {
		aspectRatio = "other"
	}
	return aspectRatio, nil

}
