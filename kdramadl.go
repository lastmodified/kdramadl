// Copyright (C) 2017 github.com/lastmodified
//
// This file is part of kdramadl.
//
// kdramadl is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// kdramadl is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with kdramadl.  If not, see <http://www.gnu.org/licenses/>.
//

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
)

const version = "0.0.1"
const formatMKV = "mkv"
const formatMP4 = "mp4"

var formats = []string{formatMKV, formatMP4}
var resolutions = []string{"1080p", "720p", "480p", "360p"}

var progHeader = fmt.Sprintf(
	`=====================================================
KDRAMA DOWNLOADER (%v)
=====================================================
`, version)
var userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:10.0) Gecko/20150101 Firefox/47.0 (Chrome)"

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func genFfmpegCmd(
	ffmpegPath string, ffmpegLogLevel string, timeout int,
	vidURL string, subURL string,
	format string, partFilePath string) *exec.Cmd {
	args := []string{"-loglevel", ffmpegLogLevel, "-stats", "-y",
		"-timeout", fmt.Sprintf("%v", timeout*1000000), // in microseconds
		"-rw_timeout", fmt.Sprintf("%v", timeout*1000000), // in microseconds
		"-reconnect", "1", "-reconnect_streamed", "1",
		"-i", vidURL, "-i", subURL}
	if format == formatMP4 {
		args = append(
			args, []string{"-c:s", "mov_text", "-c:v", "libx264", "-c:a", "copy"}...)
	}
	ffmpegOutputFormat := "mp4"
	if format == formatMKV {
		args = append(
			args, []string{"-c", "copy"}...)
		ffmpegOutputFormat = "matroska"
	}
	args = append(
		args, []string{"-bsf:a", "aac_adtstoasc", "-f", ffmpegOutputFormat, partFilePath}...)
	ffmpegCmd := exec.Command(ffmpegPath, args...)
	ffmpegCmd.Stderr = os.Stderr
	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stdin = os.Stdin
	return ffmpegCmd
}

func input(promptText string, reader *bufio.Reader) string {
	fmt.Print(promptText)
	response, _ := reader.ReadString('\n')
	response = strings.Trim(response, " \n")
	return response
}

func prnToScreen(text string) {
	if strings.HasSuffix(text, "\n") {
		fmt.Printf(text)
	} else {
		fmt.Println(text)
	}
}

func prnInfo(text string) {
	text = fmt.Sprintf("[i] %v", text)
	prnToScreen(text)
}

func main() {

	var (
		dlCode     string
		res        string
		format     string
		fileName   string
		subOnly    bool
		ffmpegPath string
		dlFolder   string
		altHost    bool
		timeout    int
	)
	reader := bufio.NewReader(os.Stdin)

	app := cli.NewApp()
	app.Name = "kdramadl"
	app.Version = version
	app.Copyright = "2017 https://github.com/lastmodified/"
	app.Usage = "Alternative downloader for https://kdrama.anontpp.com"
	app.Description = "Make sure you have ffmpeg installed in PATH or in the current folder."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c, code",
			Usage:       "Download Code",
			Destination: &dlCode,
		},
		cli.StringFlag{
			Name: "r, resolution",
			Usage: fmt.Sprintf(
				"Resolution of video. Choose from: \"%v\". Default is \"%v\".",
				strings.Join(resolutions, "\" \""),
				resolutions[0]),
			Destination: &res,
		},
		cli.StringFlag{
			Name: "f, format",
			Usage: fmt.Sprintf(
				"Video format. Choose from: \"%v\". Default is \"%v\".",
				strings.Join(formats, "\" \""),
				formats[0]),
			Destination: &format,
		},
		cli.StringFlag{
			Name:        "filename",
			Usage:       "Filename to save as (without extension).",
			Destination: &fileName,
		},
		cli.BoolFlag{
			Name:        "sub",
			Usage:       "Download only subtitles.",
			Destination: &subOnly,
		},
		cli.StringFlag{
			Name:        "ffmpeg",
			Value:       "ffmpeg",
			Usage:       "Path to ffmpeg executable.",
			Destination: &ffmpegPath,
		},
		cli.StringFlag{
			Name:        "folder",
			Value:       "",
			Usage:       "Path to download folder.",
			Destination: &dlFolder,
		},
		cli.BoolFlag{
			Name:        "alt",
			Usage:       "Use kdrama.armsasuncion.com instead of kdrama.anontpp.com",
			Destination: &altHost,
		},
		cli.IntFlag{
			Name:        "timeout",
			Value:       10,
			Usage:       "Connection timeout interval in seconds. Default 10.",
			Destination: &timeout,
		},
	}
	app.Action = func(c *cli.Context) error {

		fmt.Printf(progHeader)

		if dlCode == "" {
			dlCode = input("Enter the Download Code: ", reader)
		}
		if dlCode == "" {
			return errors.New("Download Code cannot be blank")
		}
		if fileName == "" {
			fileName = input("Enter the Filename (no extension): ", reader)
		}
		if fileName == "" {
			return errors.New("Filename cannot be blank")
		}
		if res == "" {
			res = input(fmt.Sprintf(
				"Choose a Resolution (%v). Press ENTER to use the default (%v): ",
				strings.Join(resolutions, ", "), resolutions[0]), reader)
			if res == "" {
				res = resolutions[0]
			} else if stringInSlice(res, resolutions) != true {
				return fmt.Errorf("Invalid resolution: %v", res)
			}
		}
		if format == "" {
			format = input(fmt.Sprintf(
				"Choose a Format (%v). Press ENTER to use the default (%v): ",
				strings.Join(formats, ", "), formats[0]), reader)
			if format == "" {
				format = formats[0]
			} else if stringInSlice(format, formats) != true {
				return fmt.Errorf("Invalid format: %v", format)
			}
		}

		hostname := "https://kdrama.anontpp.com/"
		if altHost == true {
			hostname = "https://kdrama.armsasuncion.com/"
		}
		subURL := fmt.Sprintf(
			"%v?dcode=%v&downloadccsub=1", hostname, url.QueryEscape(dlCode))
		vidURL := fmt.Sprintf(
			"%v?dcode=%v&quality=%v&downloadmp4vid=1", hostname,
			url.QueryEscape(dlCode), url.QueryEscape(res))

		var absFolderPath string
		// cwd, _ := os.Getwd()

		ex, _ := os.Executable()
		cwd := filepath.Dir(ex)

		if dlFolder != "" {
			absFolderPath, _ = filepath.Abs(dlFolder)
			if stat, err := os.Stat(absFolderPath); err != nil || !stat.IsDir() {
				os.MkdirAll(dlFolder, os.ModePerm)
				prnInfo(fmt.Sprintf("Created folder: %v", absFolderPath))
			}
		}
		if absFolderPath == "" {
			// default to current working folder
			absFolderPath = cwd
		}
		subFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.srt", fileName))
		vidFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.%v", fileName, format))
		partFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.%v.part", fileName, format))

		httpClient := &http.Client{}

		// Download subtitles
		if subOnly == true || format == formatMP4 {

			request, _ := http.NewRequest("GET", subURL, nil)
			request.Header.Set("User-Agent", userAgent)
			response, err := httpClient.Do(request)
			if err != nil {
				return fmt.Errorf("Error downloading subtitles: %v", err)
			}
			if response.StatusCode >= 400 {
				return fmt.Errorf("Error downloading subtitles: HTTP %v", response.StatusCode)
			}
			contentType := response.Header.Get("content-type")
			if strings.Contains(contentType, "text/html") {
				return fmt.Errorf(
					"Error downloading subtitles: Unexpected Content-Type \"%v\"",
					contentType)
			}
			defer response.Body.Close()
			output, err := os.Create(subFilePath)
			if err != nil {
				return fmt.Errorf("%v already exists", subFilePath)
			}
			defer output.Close()
			if _, err := io.Copy(output, response.Body); err != nil {
				return fmt.Errorf("Error downloading subtitles: %v", err)
			}
			prnInfo(fmt.Sprintf("Saved subtitles: %v", subFilePath))
		}
		if subOnly == true {
			return nil
		}

		// List of potential ffmpeg paths
		ffmpegPaths := []string{
			ffmpegPath, path.Join(cwd, "ffmpeg"), path.Join(cwd, "ffmpeg.exe")}

		var verifiedFfmpegPath string
		for i := 0; i < len(ffmpegPaths); i++ {
			ffmpegCmd := exec.Command(ffmpegPaths[i], "-version")
			err := ffmpegCmd.Run()
			if err == nil {
				verifiedFfmpegPath = ffmpegPaths[i]
				break
			}
		}
		if verifiedFfmpegPath == "" {
			// no ffmpeg found
			return errors.New("Unable to find valid ffmpeg path")
		}
		ffmpegLogLevel := "fatal"
		ffmpegCmd := genFfmpegCmd(
			verifiedFfmpegPath, ffmpegLogLevel, timeout, vidURL, subURL, format, partFilePath)

		if err := ffmpegCmd.Run(); err != nil {
			// Retry with a more verbose loglevel
			ffmpegCmd := genFfmpegCmd(
				verifiedFfmpegPath, "warning", timeout, vidURL, subURL, format, partFilePath)
			err := ffmpegCmd.Run()
			if err != nil {
				// Do http request to check what's wrong
				request, _ := http.NewRequest("GET", subURL, nil)
				request.Header.Set("User-Agent", userAgent)
				response, err := httpClient.Do(request)

				if err != nil {
					return fmt.Errorf("Error downloading video: %v", err)
				}
				if response.StatusCode >= 400 {
					return fmt.Errorf("Error downloading video: HTTP %v", response.StatusCode)
				}
				contentType := response.Header.Get("content-type")
				if strings.Contains(contentType, "text/html") {
					return fmt.Errorf(
						"Error downloading video: Unexpected Content-Type \"%v\"",
						contentType)
				}
				return fmt.Errorf("ffmpeg Error: %v", err)
			}
		}
		if _, err := os.Stat(partFilePath); !os.IsNotExist(err) {
			// rename .part file to final filename
			err := os.Rename(partFilePath, vidFilePath)
			if err != nil {
				fmt.Printf("Unable to rename file %v to %v\n", partFilePath, vidFilePath)
				return errors.New("Unable to rename file")
			}
		}
		if _, err := os.Stat(vidFilePath); !os.IsNotExist(err) {
			prnInfo(fmt.Sprintf("Saved video %v", vidFilePath))
		}
		input("\bPress ENTER to continue...", reader)
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("\n[x] %v\n\n", err)
		input("\bPress ENTER to continue...", reader)
	}
}
