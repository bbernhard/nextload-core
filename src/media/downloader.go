package media

import (
	"os/exec"
	"bufio"
	"errors"
	log "github.com/Sirupsen/logrus"
)

type Downloader struct {
    pathToYoutubeDl string
    videosOutputFolder string
    audiosOutputFolder string
}

func NewDownloader(pathToYoutubeDl string, videosOutputFolder string, audiosOutputFolder string) *Downloader {
    return &Downloader {
    	pathToYoutubeDl: pathToYoutubeDl,
    	audiosOutputFolder: audiosOutputFolder,
    	videosOutputFolder: videosOutputFolder,
    } 
}

func (p *Downloader) getOutputFilename(media Media) (string, error) {
	var cmd *exec.Cmd
	if media.Format == "mp3" {
		cmd = exec.Command(p.pathToYoutubeDl, media.Url, "--get-filename", "--extract-audio", "--prefer-ffmpeg", "--audio-format", "mp3", "-o", p.audiosOutputFolder + "/%(title)s.mp3")
	} else {
		cmd = exec.Command(p.pathToYoutubeDl, media.Url, "--get-filename", "--prefer-ffmpeg",  "-o", p.videosOutputFolder + "/%(title)s.%(ext)s")
	}

	filename := ""
	stdoutCmdReader, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdoutCmdReader)
	go func() {
	    for scanner.Scan() {
	        filename += scanner.Text()
	    }
	}()

	cmdErr := "" 
	stderrCmdReader, _ := cmd.StderrPipe()
	errorScanner := bufio.NewScanner(stderrCmdReader)
	go func() {
	    for errorScanner.Scan() {
	    	cmdErr += errorScanner.Text()
	    }
	}()

	err := cmd.Start()
	if err != nil {
		return filename, err
	}
	cmd.Wait()

	if cmdErr != "" {
		return filename, errors.New(cmdErr)
	}

	return filename, nil
}

func (p *Downloader) Download(media Media) (string, error, string) {
	filename, err := p.getOutputFilename(media)
	if err != nil {
		return filename, err, ""
	}

	var cmd *exec.Cmd
	if media.Format == "mp3" {
		cmd = exec.Command(p.pathToYoutubeDl, media.Url, "--newline", "--prefer-ffmpeg", "--extract-audio", "--audio-format", "mp3", "-o", filename)
	} else {
		cmd = exec.Command(p.pathToYoutubeDl, media.Url, "--newline", "--prefer-ffmpeg", "-o", filename)
	}

	cmdOutput := ""
	stdoutCmdReader, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(stdoutCmdReader)
	go func() {
	    for scanner.Scan() {
	    	log.Debug(scanner.Text())
	        cmdOutput += scanner.Text()
	    }
	}()

	cmdErr := "" 
	stderrCmdReader, _ := cmd.StderrPipe()
	errorScanner := bufio.NewScanner(stderrCmdReader)
	go func() {
	    for errorScanner.Scan() {
	    	cmdErr += errorScanner.Text()
	    }
	}()

	err = cmd.Start()
	if err != nil {
		return filename, err, cmdOutput + "\n\n" + cmdErr
	}
	cmd.Wait()

	if cmdErr != "" {
		return filename, errors.New(cmdErr), cmdOutput + "\n\n" + cmdErr
	}

	return filename, nil, cmdOutput + "\n\n" + cmdErr
}