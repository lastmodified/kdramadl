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
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/urfave/cli"
)

const version = "0.1.5"
const formatMKV = "mkv"
const formatMP4 = "mp4"

var formats = []string{formatMKV, formatMP4}
var hostMain = "goplay.anontpp.com"
var hostAlt = "kdrama.armsasuncion.com"

var progHeader = fmt.Sprintf(
	`=====================================================
KDRAMA DOWNLOADER (%v)
=====================================================
`, version)
var userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:10.0) Gecko/20150101 Firefox/47.0 (Chrome)"
var invalidDlCodeCharRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
var validResRegex = regexp.MustCompile(`^([0-9]{3,4}p|[1-9])$`)
var logger = &custLogger{level: levelInfo}

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
		proxy      string
		timeout    int
		autoQuit   bool
		verbose    bool
	)
	reader := bufio.NewReader(os.Stdin)

	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Fprintf(
			c.App.Writer, "%v version %v\nCheck %v for the latest update.\n",
			c.App.Name, c.App.Version, "https://github.com/lastmodified/kdramadl/releases/latest")
	}
	app := cli.NewApp()
	app.Name = "kdramadl"
	app.Version = version
	app.Copyright = "2017 https://github.com/lastmodified/kdramadl"
	app.Usage = "Alternative downloader for https://goplay.anontpp.com"
	app.Description = "Make sure you have ffmpeg installed in PATH or in the current folder."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "c, code",
			Usage:       "Download Code",
			Destination: &dlCode,
		},
		cli.StringFlag{
			Name:        "r, resolution",
			Usage:       "Resolution of video, for example: 720p.",
			Destination: &res,
		},
		cli.StringFlag{
			Name: "f, format",
			Usage: fmt.Sprintf(
				"Video format. Choose from: \"%v\". Default is %q.",
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
			Usage:       fmt.Sprintf("Use %v instead of %v", hostMain, hostAlt),
			Destination: &altHost,
		},
		cli.StringFlag{
			Name:        "proxy",
			Value:       "",
			Usage:       "Proxy address (only HTTP proxies supported), example \"http://127.0.0.1:80\".",
			Destination: &proxy,
		},
		cli.IntFlag{
			Name:        "timeout",
			Value:       10,
			Usage:       "Connection timeout interval in seconds. Default 10.",
			Destination: &timeout,
		},
		cli.BoolFlag{
			Name:        "autoquit",
			Usage:       "Automatically quit when done (skip the \"Press ENTER to continue\" prompt)",
			Destination: &autoQuit,
		},
		cli.BoolFlag{
			Name:  "nocolor",
			Usage: "Disable color output",
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "Generate more verbose messages",
			Destination: &verbose,
		},
	}
	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		if isSubcommand {
			return err
		}

		cli.ShowAppHelp(c)
		fmt.Fprintf(c.App.Writer, "\nUsage error: %v\n", err)
		return nil
	}
	app.Action = func(c *cli.Context) error {

		if c.Bool("nocolor") {
			color.NoColor = true
		}
		if verbose {
			logger.level = levelDebug
		}
		fmt.Printf(progHeader)

		ex, _ := os.Executable()
		cwd := filepath.Dir(ex)
		// List of potential ffmpeg paths
		ffmpegPaths := []string{
			ffmpegPath, path.Join(cwd, "ffmpeg"), path.Join(cwd, "ffmpeg.exe")}

		var verifiedFfmpegPath string
		for _, testPath := range ffmpegPaths {
			ffmpegCmd := exec.Command(testPath, "-version")
			err := ffmpegCmd.Run()
			if err == nil {
				verifiedFfmpegPath = testPath
				break
			}
		}
		if verifiedFfmpegPath == "" {
			// no ffmpeg found
			return errors.New("Unable to find valid ffmpeg path")
		}

		var httpClient *http.Client
		if proxy == "" {
			httpClient = &http.Client{}
		} else {
			proxyURL, err := url.Parse(proxy)
			if err != nil {
				return err
			}
			if !strings.HasPrefix(proxyURL.Scheme, "http") {
				// Because ffmpeg does not support SOCKS proxies
				return fmt.Errorf("Unsupport proxy scheme: %v", proxyURL.Scheme)
			}
			httpClient = &http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
			}
		}

		if dlCode == "" {
			dlCode = input("Enter the Download Code: ", reader)
		}
		if dlCode == "" {
			return errors.New("Download Code cannot be blank")
		} else if invalidDlCodeCharRegex.MatchString(dlCode) {
			return errors.New("Invalid Download Code")
		}
		if fileName == "" {
			fileName = input("Enter the Filename (no extension): ", reader)
		}
		if fileName == "" {
			return errors.New("Filename cannot be blank")
		}

		if res == "" {
			res = input("Enter a Resolution (please check on video page): ", reader)
		}
		if res == "" {
			return errors.New("Resolution cannot be blank")
		} else if validResRegex.MatchString(res) != true {
			return fmt.Errorf("Invalid resolution: %v", res)
		}

		if format == "" {
			format = input(fmt.Sprintf(
				"Choose a Format (%v). Press ENTER to use the default (%v): ",
				strings.Join(formats, ", "), formats[0]), reader)
		}
		if format == "" {
			format = formats[0]
		} else if stringInSlice(format, formats) != true {
			return fmt.Errorf("Invalid format: %v", format)
		}

		hostname := fmt.Sprintf("https://%v/", hostMain)
		if altHost == true {
			hostname = fmt.Sprintf("https://%v/", hostAlt)
		}
		subURL := fmt.Sprintf(
			"%v?dcode=%v&downloadccsub=1", hostname, url.QueryEscape(dlCode))
		vidURL := fmt.Sprintf(
			"%v?dcode=%v&quality=%v&downloadmp4vid=1", hostname,
			url.QueryEscape(dlCode), url.QueryEscape(res))

		var absFolderPath string
		if dlFolder != "" {
			absFolderPath, _ = filepath.Abs(dlFolder)
			if stat, err := os.Stat(absFolderPath); err != nil || !stat.IsDir() {
				os.MkdirAll(dlFolder, os.ModePerm)
				logger.Infof("Created folder: %v", absFolderPath)
			}
		}
		if absFolderPath == "" {
			// default to current working folder
			absFolderPath = cwd
		}
		logger.Debugf(
			"Download Code: %v, Resolution: %v, Filename: %v, Format: %v, Folder: %v",
			dlCode, res, fileName, format, absFolderPath)
		subFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.srt", fileName))
		vidFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.%v", fileName, format))
		partFilePath := path.Join(absFolderPath, fmt.Sprintf("%v.%v.part", fileName, format))

		// Download subtitles
		if subOnly == true || format == formatMP4 {

			request, _ := http.NewRequest("GET", subURL, nil)
			request.Header.Set("User-Agent", userAgent)
			logger.Debugf("Requesting %v", subURL)
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
			logger.Infof("Saved subtitles: %v", subFilePath)
		}
		if subOnly == true {
			return nil
		}

		ffmpegLogLevel := "fatal"
		if verbose {
			ffmpegLogLevel = "info"
		}
		ffmpegCmd := genFfmpegCmd(
			verifiedFfmpegPath, ffmpegLogLevel, timeout,
			vidURL, subURL, format, partFilePath, proxy)
		logger.Debugf("Requesting %v", vidURL)
		logger.Debugf("FFMPEG args: %v", ffmpegCmd.Args)

		if err := ffmpegCmd.Run(); err != nil {
			// Retry with a more verbose loglevel
			if ffmpegLogLevel == "fatal" {
				ffmpegLogLevel = "warning"
			}
			ffmpegCmd := genFfmpegCmd(
				verifiedFfmpegPath, ffmpegLogLevel, timeout,
				vidURL, subURL, format, partFilePath, proxy)
			logger.Debugf("Requesting %v", vidURL)
			logger.Debugf("FFMPEG args: %v", ffmpegCmd.Args)
			err := ffmpegCmd.Run()
			if err != nil {
				// Do http request to check what's wrong
				request, _ := http.NewRequest("GET", vidURL, nil)
				request.Header.Set("User-Agent", userAgent)
				response, err := httpClient.Do(request)

				if err != nil {
					return fmt.Errorf("Error downloading video: %v", err)
				}
				if response.StatusCode >= 400 {
					return fmt.Errorf(
						"Error downloading video: HTTP %v %q",
						response.StatusCode, response.Request.URL.String())
				}
				contentType := response.Header.Get("content-type")
				if strings.Contains(contentType, "text/html") {
					return fmt.Errorf(
						"Error downloading video: Unexpected Content-Type %q",
						contentType)
				}
				return fmt.Errorf("ffmpeg Error: %v", err)
			}
		}
		if _, err := os.Stat(partFilePath); !os.IsNotExist(err) {
			// rename .part file to final filename
			err := os.Rename(partFilePath, vidFilePath)
			if err != nil {
				logger.Debugf("Error renaming %q to %q: %v", partFilePath, vidFilePath, err)
				return errors.New("Unable to rename file")
			}
		}
		if _, err := os.Stat(vidFilePath); !os.IsNotExist(err) {
			logger.Infof("Saved video: %v", vidFilePath)
		}
		if !autoQuit {
			input("\bPress ENTER to continue...", reader)
		}
		return nil
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Errorf("%v", err)
		if !autoQuit {
			input("\bPress ENTER to continue...", reader)
		}
	}
}

func genFfmpegCmd(
	ffmpegPath string, ffmpegLogLevel string, timeout int,
	vidURL string, subURL string,
	format string, partFilePath string, proxy string) *exec.Cmd {
	args := []string{"-loglevel", ffmpegLogLevel, "-stats", "-y",
		"-timeout", fmt.Sprintf("%v", timeout*1000000), // in microseconds
		"-reconnect", "1", "-reconnect_streamed", "1"}
	if proxy != "" {
		args = append(args, []string{"-http_proxy", proxy}...)
	}
	args = append(args, []string{"-i", vidURL, "-i", subURL}...)
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

const (
	levelCritical = 50
	levelError    = 40
	levelWarning  = 30
	levelInfo     = 20
	levelDebug    = 10
	levelNoSet    = 0
)

// Very basic logger with levels
type custLogger struct {
	level int
}

var red = color.New(color.FgRed).SprintFunc()
var yellow = color.New(color.FgYellow).Add(color.BgBlack).SprintFunc()
var blue = color.New(color.FgBlue).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()
var bold = color.New(color.Bold).SprintFunc()

func (logger custLogger) Log(level int, message string) {
	switch level {
	case levelCritical:
		message = fmt.Sprintf("%v: %v", red("CRITICAL"), message)
	case levelError:
		message = fmt.Sprintf("%v: %v", red("ERROR"), message)
	case levelWarning:
		message = fmt.Sprintf("%v: %v", yellow("WARNING"), message)
	case levelInfo:
		message = fmt.Sprintf("%v: %v", green("INFO"), message)
	case levelDebug:
		message = fmt.Sprintf("%v: %v", blue("DEBUG"), message)
	}

	if level >= logger.level {
		if strings.HasSuffix(message, "\n") {
			fmt.Print(message)
		} else {
			fmt.Println(message)
		}
	}
}
func (logger custLogger) Logf(level int, msg string, a ...interface{}) {
	logger.Log(level, fmt.Sprintf(msg, a...))
}
func (logger custLogger) Debug(msg string)                       { logger.Log(levelDebug, msg) }
func (logger custLogger) Info(msg string)                        { logger.Log(levelInfo, msg) }
func (logger custLogger) Warning(msg string)                     { logger.Log(levelWarning, msg) }
func (logger custLogger) Error(msg string)                       { logger.Log(levelError, msg) }
func (logger custLogger) Critical(msg string)                    { logger.Log(levelCritical, msg) }
func (logger custLogger) Debugf(msg string, a ...interface{})    { logger.Logf(levelDebug, msg, a...) }
func (logger custLogger) Infof(msg string, a ...interface{})     { logger.Logf(levelInfo, msg, a...) }
func (logger custLogger) Warningf(msg string, a ...interface{})  { logger.Logf(levelWarning, msg, a...) }
func (logger custLogger) Errorf(msg string, a ...interface{})    { logger.Logf(levelError, msg, a...) }
func (logger custLogger) Criticalf(msg string, a ...interface{}) { logger.Logf(levelDebug, msg, a...) }
