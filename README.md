nextload-core is a small dockerized service, that downloads videos from Youtube and a few other [sites](https://rg3.github.io/youtube-dl/supportedsites.html) and then uploads them to your nextcloud instance.

# Requirements

The only requirements are: 

* docker
* docker-compose

You can run nextload-core on the same host as nextcloud, but you don't need to. nextload-core can be installed on every system that supports docker and docker-compose. 

# Install

As nextload-core accesses the nextcloud instance via webdav, we first need to create a new app token. 

* Navigate to Profile -> Settings -> Security to create a new app token

Name your token `nextload` and click "Generate new app password" to generate a new token. Now copy the generate token to your clipboard...we will need that one a bit later.




* download the `docker-compose.yml` file with `...` to your host system
* open the `docker-compose.yml` file with an editor and set the `NEXTLOAD_URL`, `NEXTLOAD_TOKEN` parameter accordingly. 
  If everything is set correctly, the file should look like this: 

  ```
  version: '3'
  services:
    nextload:
      image: nextload-core:latest
      restart: always
      environment:
        - NEXTCLOUD_TOKEN=ZgASA-cRxSg-HASAA-sd5Gz-qCtyr
        - NEXTCLOUD_URL=https://cloud.example.com
        - POLL_INTERVAL=5
      volumes:
        - ./logs:/var/log/nextload-core
  ```
* build and run docker-compose file with: `docker-compose up --pull` resp. ``docker-compose up --pull -d` if you want to start container in detached mode.


# How it works

When the docker container starts up, it will create a new `nextload` folder in your home directory. 

[image]

Inside this folder, there are the nextload config file (`config.yml`) and two folders (`audios`, `videos`). 


[image]

If you want to create a new download task, just create a new `.yml` file inside this directory. The actual name of the file doesn't matter, it just needs to have the ending `.yml`. 

[ image ]

In the text editor that opens, specify the url and the download format: 

[image ]

Per default, the docker container polls the nextcloud instance every 5 minutes (the `POLL_INTERVAL` can be changed in the `docker-compose.yml` file) for new download tasks. In case a new download task appeared, nextload-core downloads the file and then uploads it to your nextcloud account. 

So, after 5+ minutes you should see the file appear either in the `audios` or the `videos` folder (depending on the format you specified).

[image]

In case a download task couldn't be processed, nextload-core creates a new file called `task-file-name.error.txt` in your nextcloud instance containing the error message. So, if e.q: the task `1.yml` couldn't be processed, you will see the error log in the file `1.error.txt`. 





