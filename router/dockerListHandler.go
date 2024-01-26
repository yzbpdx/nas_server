package router

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"nas_server/logs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DockerListHandler(ctx *gin.Context, port string) {
	if !CookieVerify(ctx) {
		return
	}

	host := ctx.Request.Host
	addr := strings.Split(host, ":")
	domain := "http://" + addr[0] + ":" + port

	repoResp, _ := http.Get(domain + "/v2/_catalog")
	defer repoResp.Body.Close()
	repoRespBody, _ := io.ReadAll(repoResp.Body)
	var repos struct {
		Repos []string `json:"repositories"`
	}
	if err := json.Unmarshal(repoRespBody, &repos); err != nil {
		logs.GetInstance().Logger.Errorf("unmarshal docker repos error %s", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	logs.GetInstance().Logger.Infof("find docker repos %v", repos)

	result := make([]DockerRepositry, 0, len(repos.Repos))
	for _, repo := range repos.Repos {
		tagUrl := fmt.Sprintf(domain + "/v2/%s/tags/list", repo)

		tagResp, _ := http.Get(tagUrl)
		defer tagResp.Body.Close()

		tagRespBody, _ := io.ReadAll(tagResp.Body)
		var tags struct {
			Tags []string `json:"tags"`
		}
		if err := json.Unmarshal(tagRespBody, &tags); err != nil {
			logs.GetInstance().Logger.Errorf("unmarshal docker tags error %s", err)
			// ctx.JSON(http.StatusInternalServerError, gin.H{})
			continue
		}

		for _, tag := range tags.Tags {
			manifestUrl := fmt.Sprintf(domain + "/v2/%s/manifests/%s", repo, tag)

			manifestReq, _ := http.NewRequest("GET", manifestUrl, nil)
			manifestReq.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

			client := &http.Client{}
			manifestResp, _ := client.Do(manifestReq)
			defer manifestResp.Body.Close()

			manifestRespBody, _ := io.ReadAll(manifestResp.Body)
			var manifest struct {
				Config struct {
					Labels map[string]string `json:"labels"`
				} `json:"config"`
				Layers []struct {
					Digest string `json:"digest"`
					Size   int64  `json:"size"`
				} `json:"layers"`
			}
			if err := json.Unmarshal(manifestRespBody, &manifest); err != nil {
				logs.GetInstance().Logger.Errorf("unmarshal docker manifest error %s", err)
				// ctx.JSON(http.StatusInternalServerError, gin.H{})
				continue
			}

			dockerRepositry := DockerRepositry {
				Repository: repo,
				Tag: tag,
				ImageId: strings.TrimPrefix(manifest.Config.Labels["org.opencontainers.image.ref.name"], "sha256:"),
				Created: manifest.Config.Labels["org.opencontainers.image.created"],
				Pull: fmt.Sprintf("docker pull %s/%s:%s", domain, repo, tag),
			}
			size := int64(0)
			for _, layer := range manifest.Layers {
				size += layer.Size
			}
			fmt.Println(float64(size) / math.Pow(1024, 3))
			if float64(size) / math.Pow(1024, 3) >= 1 {
				dockerRepositry.Size = fmt.Sprintf("%.2fGB", float64(size) / math.Pow(1024, 3))
			} else {
				dockerRepositry.Size = fmt.Sprintf("%.2fMB", float64(size) / math.Pow(1024, 2))
			}
			result = append(result, dockerRepositry)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"dockers": result,
	})
}