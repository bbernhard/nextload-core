package main

import (
	"flag"
	"path/filepath"
	"strings"
	"os"
	"time"
	"strconv"
	log "github.com/Sirupsen/logrus"
	nextcloud "./nextcloud"
	misc "./misc"
	media "./media"
	config "./config"
)



func main() {
	nextCloudUrlParam := flag.String("nextcloud-url", "", "Specify the nextcloud URL")
	nextCloudTokenParam := flag.String("nextcloud-token", "", "Specify the device token")
	releaseMode := flag.Bool("release", false, "Run in release mode")
	youtubeDl := flag.String("youtube-dl", "../ext/windows/youtube-dl/youtube-dl.exe", "Specify the path to youtube dl")
	pollIntervalParam := flag.Int("poll-interval", 5, "Specify the poll interval (in minutes)")
	tempVideosDownloadFolder := flag.String("temp-videos-download-folder", "", "/tmp/videos/")
	tempAudiosDownloadFolder := flag.String("temp-audios-download-folder", "", "/tmp/audios/")
	defaultConfigFilePath := flag.String("default-config-file", "../config/config.default.yml", "Path to the default config file")

	flag.Parse()

	nextCloudUrl := *nextCloudUrlParam
	if *nextCloudUrlParam == "" {
		nextCloudUrl = os.Getenv("NEXTCLOUD_URL")
	}

	if nextCloudUrl == "" {
		log.Fatal("[Main] Please provide a valid nextcloud url")
	}


	nextCloudToken := *nextCloudTokenParam
	if *nextCloudTokenParam == "" {
		nextCloudToken = os.Getenv("NEXTCLOUD_TOKEN")
	}

	pollInterval := *pollIntervalParam
	if os.Getenv("POLL_INTERVAL") != "" {
		interval, err := strconv.ParseInt(os.Getenv("POLL_INTERVAL"), 10, 32)
		if err != nil {
			log.Fatal("[Main] Please provide a valid poll interval")
		}

		pollInterval = int(interval)
	}

	if nextCloudToken == "" {
		log.Fatal("[Main] Please provide a valid nextcloud token")
	}

	log.Info("[Main] Starting daemon")

	if ! *releaseMode {
		log.SetLevel(log.DebugLevel)
	}

	folderName := "nextload"

	nextCloudClient := nextcloud.NewNextCloudClient(nextCloudUrl, nextCloudToken)
	folderExists, err := nextCloudClient.FolderExists(folderName)
	if err != nil {
		log.Fatal("[Main] Couldn't check whether folder exists: ", err.Error())
	}

	if !folderExists {
		log.Info("[Main] Creating folder with name ", folderName)
		err = nextCloudClient.CreateFolder(folderName)
		if err != nil {
			log.Fatal("[Main] Couldn't create folder: ", err.Error())
		}
	}

	configFileExists, err := nextCloudClient.FileExists(folderName + "/config.yml")
	if err != nil {
		log.Fatal("[Main] Couldn't check whether config exists: ", err.Error())
	}

	if !configFileExists {
		log.Info("[Main] Config doesn't exist...creating default config")

		err = nextCloudClient.Upload(*defaultConfigFilePath, folderName + "/config.yml")
		if err != nil {
			log.Fatal("[Main] Couldn't create default config")
		}
	}

	defaultCfg, err := config.GetDefaultConfig(*defaultConfigFilePath)
	if err != nil {
		log.Fatal("[Main] Couldn't get default config: ", err)
	}



	for {
		startTime := time.Now()

		configFileBytes, err := nextCloudClient.GetFile(folderName + "/config.yml")
		if err != nil {
			log.Fatal("[Main] Couldn't get config")
		}

		cfg, err := config.FromBytes(configFileBytes)
		if err != nil {
			log.Fatal("[Main] Couldn't parse config: ", err.Error())
		}

		//check whether we need to create default videos folders
		if defaultCfg.Output.Paths.Videos == cfg.Output.Paths.Videos {
			defaultFolderExists, err := nextCloudClient.FolderExists(defaultCfg.Output.Paths.Videos)
			if err != nil {
				log.Fatal("[Main] Couldn't check whether to create default videos folder: ", err.Error())
			}

			if !defaultFolderExists {
				log.Info("Creating default videos folder")
				err = nextCloudClient.CreateFolder(defaultCfg.Output.Paths.Videos)
				if err != nil {
					log.Fatal("[Main] Couldn't create default videos folder: ", err.Error())
				}
			}
		}

		//check whether we need to create default audios folders
		if defaultCfg.Output.Paths.Audios == cfg.Output.Paths.Audios {
			defaultFolderExists, err := nextCloudClient.FolderExists(defaultCfg.Output.Paths.Audios)
			if err != nil {
				log.Fatal("[Main] Couldn't check whether to create default audios folder: ", err.Error())
			}

			if !defaultFolderExists {
				log.Info("Creating default audios folder")
				err = nextCloudClient.CreateFolder(defaultCfg.Output.Paths.Audios)
				if err != nil {
					log.Fatal("[Main] Couldn't create default audios folder: ", err.Error())
				}
			}
		}


		log.Info("[Main] Checking whether config is valid")
		err = config.IsValid(cfg, nextCloudClient)
		if err != nil {
			log.Fatal("[Main] Invalid config: ", err.Error())
		}
		log.Info("[Main] Config is valid")

		downloader := media.NewDownloader(*youtubeDl, *tempVideosDownloadFolder, *tempAudiosDownloadFolder)


		contents, err := nextCloudClient.ListFolderContents(folderName)
		if err != nil {
			log.Error("Couldn't list ", err.Error())
		}



		for _, content := range contents {
			if content.ContentType != "application/yaml" {
				continue
			}

			//skip config.yml
			if content.Path == folderName + "/config.yml" {
				continue
			}

			bytes, err := nextCloudClient.GetFile(content.Path)
			if err != nil {
				log.Error("[Main] Couldn't get file: ", err.Error())
			}

			task, err := misc.ToTask(bytes)
			if err != nil {
				log.Error("[Main] Couldn't get task: ", err.Error())
			}

			log.Info("[Main] Got a new task: ", task.Url)

			m := media.Media{Url: task.Url, Format: task.Format}
			mediaPath, err, errorLog := downloader.Download(m)
			if err != nil {
				log.Error("[Main] Couldn't download media: ", err.Error())

				//upload error log file
				fname := strings.TrimSuffix(content.Path, filepath.Ext(content.Path)) + ".error.txt"
				err := nextCloudClient.UploadSerializedFile([]byte(errorLog), fname)
				if err != nil {
					log.Error("[Main] Couldn't upload error log: ", err.Error())
				}
			} else { //download was successful
				//upload media to nextcloud
				mediaUploadFolder := ""
				if media.GetTypeFromFormat(task.Format) == media.Audio {
					mediaUploadFolder = cfg.Output.Paths.Audios
				} else {
					mediaUploadFolder = cfg.Output.Paths.Videos
				}
				mediaUploadPath := mediaUploadFolder + "/" + filepath.Base(mediaPath)
				err = nextCloudClient.Upload(mediaPath, mediaUploadPath)
				if err != nil {
					log.Error("[Main] Couldn't upload media: ", mediaUploadPath, " ", err.Error())
				}
				log.Info("[Main] Successfully uploaded file to ", mediaUploadPath)

				//remove task
				log.Info("[Main] Removing task ", content.Path)
				err = nextCloudClient.RemoveFile(content.Path)
				if err != nil {
					log.Error("[Main] Couldn't remove task: ", content.Path, " ", err.Error())
				}

				//remove any existing error log for that task
				//we do not need to check whether it succeeds or not, we can't do anything about it anyway
				log.Info("[Main] Removing error log for task ", content.Path)
				nextCloudClient.RemoveFile(strings.TrimSuffix(content.Path, filepath.Ext(content.Path)) + ".error.txt")
			}

			log.Info("[Main] Removing ", mediaPath)
			os.Remove(mediaPath) //no need to check if successful. we can't do anything about it anyway
		}

		endTime := time.Now()

		timeDiff := endTime.Sub(startTime) 
		if ((int64(pollInterval * 6e+10) - int64(timeDiff)) < 0) {
			continue //time already passed, loop immediatelly again
		}

		time.Sleep(time.Duration(int64(pollInterval * 6e+10) - int64(timeDiff))) //sleep for x seconds
	}
}