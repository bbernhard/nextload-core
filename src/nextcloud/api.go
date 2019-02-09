package nextcloud

import (
	"io/ioutil"
	"gopkg.in/resty.v1"
	"net/http"
	//log "github.com/Sirupsen/logrus"
	"errors"
	"github.com/beevik/etree"
	"strings"
)

type Item struct {
	Name string
	ContentType string
	Path string
}

type NextCloudClient struct {
    url string
    token string
    timeoutInSecs int
}

func NewNextCloudClient(url string, token string) *NextCloudClient {
    return &NextCloudClient {
    	url: url,
        token: token,
        timeoutInSecs: 10,
    } 
}

func (p *NextCloudClient) Upload(source string, dest string) error {
	u := p.url + "/remote.php/webdav/" + dest

    fileBytes, err := ioutil.ReadFile(source)
    if err != nil {
    	return err
    }

    _, err = resty.R().
		SetBody(fileBytes).
		SetContentLength(true).
		SetAuthToken(p.token).
		Put(u)

    return err
}

func (p *NextCloudClient) UploadSerializedFile(fileBytes []byte, dest string) error {
	u := p.url + "/remote.php/webdav/" + dest

    _, err := resty.R().
		SetBody(fileBytes).
		SetContentLength(true).
		SetAuthToken(p.token).
		Put(u)

    return err
}


func (p *NextCloudClient) FolderExists(name string) (bool, error) {
	u := p.url + "/remote.php/webdav/" + name

	req, err := http.NewRequest("PROPFIND", u, nil)
	if err != nil {
		return false, err
	}
	
	req.Header.Add("Authorization", "Bearer " + p.token)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 207 {
		return true, nil
	}

	return false, nil
}

func (p *NextCloudClient) FileExists(name string) (bool, error) {
	u := p.url + "/remote.php/webdav/" + name

	req, err := http.NewRequest("PROPFIND", u, nil)
	if err != nil {
		return false, err
	}
	
	req.Header.Add("Authorization", "Bearer " + p.token)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	if res.StatusCode == 207 {
		return true, nil
	}

	return false, nil
}

func (p *NextCloudClient) RemoveFile(name string) error {
	u := p.url + "/remote.php/webdav/" + name

	req, err := http.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}
	
	req.Header.Add("Authorization", "Bearer " + p.token)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 200 || res.StatusCode == 201 || res.StatusCode == 202 || res.StatusCode == 204  {
		return nil
	}

    return errors.New("Couldn't remove file")
}

func (p *NextCloudClient) CreateFolder(name string) error {
	u := p.url + "/remote.php/webdav/" + name

	req, err := http.NewRequest("MKCOL", u, nil)
	if err != nil {
		return err
	}
	
	req.Header.Add("Authorization", "Bearer " + p.token)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == 201 {
		return nil
	} else if res.StatusCode == 405 { //folder already exists
		return errors.New("Folder already exists or no permissions")
	}

	return errors.New("Unknown error")
}

func (p *NextCloudClient) ListFolderContents(folder string) ([]Item, error) {
	tasks := []Item{}

	u := p.url + "/remote.php/webdav/" + folder

	req, err := http.NewRequest("PROPFIND", u, nil)
	if err != nil {
		return tasks, err
	}
	
	req.Header.Add("Authorization", "Bearer " + p.token)

	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return tasks, err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return tasks, err
	}

	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(respBody); err != nil {
		return tasks, err
	}

	root := doc.SelectElement("d:multistatus")
	for _, c := range root.SelectElements("d:response") {
		cType := ""
		if propStat := c.SelectElement("d:propstat"); propStat != nil {
			if prop := propStat.SelectElement("d:prop"); prop != nil {
				if contentType := prop.SelectElement("d:getcontenttype"); contentType != nil {
					cType = contentType.Text()
				}
			}
		}

		if cType != "application/yaml" {
			continue
		}

		if url := c.SelectElement("d:href"); url != nil {
			f := strings.TrimPrefix(url.Text(), "/remote.php/webdav/" + strings.Trim(folder, "/") + "/")
			f = strings.Trim(f, "/")
			if f == "" {
				continue
			}

			path := strings.Trim(folder, "/") + "/" + f

			item := Item{ContentType: cType, Name: f, Path: path}
			tasks = append(tasks, item)
		}
	}

	return tasks, nil
}

func (p *NextCloudClient) GetFile(name string) ([]byte, error) {
	u := p.url + "/remote.php/webdav/" + name

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return []byte{}, err
	}

	
	req.Header.Add("Authorization", "Bearer " + p.token)
	res, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()

	respBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []byte{}, err
	}

	return respBody, nil
}