package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go-api/common"
	"go-api/config"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type artifactRepositoryElement struct {
	Id                string `json:"repository_id"`
	Name              string `json:"name"`
	URL               string `json:"url"`
	Kind              int    `json:"kind"`
	VerifiedPublisher bool   `json:"verified_publisher"`
	Official          bool   `json:"official"`
	Disabled          bool   `json:"disabled"`
}

type artifactPackageElement struct {
	Id          string `json:"package_id"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Description string `json:"description"`
	LogoImageId string `json:"logo_image_id"`
	Icon        string `json:"icon"`
	Deprecated  bool   `json:"deprecated"`
}

type artifactPackageList struct {
	Packages []artifactPackageElement `json:"packages"`
}

type artifactPackage struct {
	Id                string        `json:"package_id"`
	Name              string        `json:"name"`
	Version           string        `json:"version"`
	AppVersion        string        `json:"app_version"`
	Description       string        `json:"description"`
	LogoImageId       string        `json:"logo_image_id"`
	Icon              string        `json:"icon"`
	Deprecated        bool          `json:"deprecated"`
	License           string        `json:"license"`
	HomeUrl           string        `json:"home_url"`
	AvailableVersions []interface{} `json:"available_versions"`
	Links             []interface{} `json:"links"`
	ContentUrl        string        `json:"content_url"`
	Repository        interface{}   `json:"repository"`
}

type respData struct {
	TotalCount int
	Data       []byte
}

// SearchRepoHub
// @Summary Search Repo ArtifactHub
// @Tags ArtifactHub
// @Accept json
// @Produce json
// @Router /api/hub/repositories [Get]
func SearchRepoHub(c *fiber.Ctx) error {
	lse, err := ListSearchCheck(c)
	if err != nil {
		return common.RespErr(c, err)
	}

	name := c.Query("name")
	url := c.Query("url")
	params := fmt.Sprintf("&limit=0&name=%v&url=%v", name, url)
	reqUrl := fmt.Sprintf("%v%v", config.Env.ArtifactHubUrl, config.Env.ArtifactHubRepoSearch) + params
	respData, err := getRequestData(reqUrl, true)
	if err != nil {
		return common.RespErr(c, err)
	}

	var repoElements []artifactRepositoryElement
	if err := json.Unmarshal(respData.Data, &repoElements); err != nil {
		return common.RespErr(c, err)
	}

	repos := make([]interface{}, 0, len(repoElements))
	for _, re := range repoElements {
		if !re.Disabled && re.VerifiedPublisher {
			repos = append(repos, re)
		}
	}

	itemCount, resultData := ResourceListProcessing(repos, lse)
	return common.ListRespOK(c, itemCount, resultData)
}

// SearchPackageHub
// @Summary Search Package ArtifactHub
// @Tags ArtifactHub
// @Accept json
// @Produce json
// @Router /api/hub/packages [Get]
func SearchPackageHub(c *fiber.Ctx) error {
	lse, err := ListSearchCheck(c)
	if err != nil {
		return common.RespErr(c, err)
	}

	if lse.Limit < 1 || lse.Limit > 60 {
		return common.RespErr(c, fmt.Errorf(common.HUB_PACKAGE_LIMIT_ILLEGAL_ARGUMENT))
	}
	repo := c.Query("repo")
	query := c.Query("query")

	params := fmt.Sprintf("&offset=%v&limit=%v&ts_query_web=%v", lse.Offset*lse.Limit, lse.Limit, query)
	if len(repo) > 0 {
		params += "&repo=" + repo
	}

	reqUrl := fmt.Sprintf("%v%v", config.Env.ArtifactHubUrl, config.Env.ArtifactHubPackageSearch) + params
	respData, err := getRequestData(reqUrl, true)
	if err != nil {
		return common.RespErr(c, err)
	}

	var artifactPackageList artifactPackageList
	if err := json.Unmarshal(respData.Data, &artifactPackageList); err != nil {
		return common.RespErr(c, err)
	}

	remainingItemCount := respData.TotalCount - ((lse.Offset + 1) * lse.Limit)
	if remainingItemCount < 0 {
		remainingItemCount = 0
	}
	listCount := common.ListCount{
		AllItemCount:       respData.TotalCount,
		RemainingItemCount: remainingItemCount,
	}

	packages := make([]interface{}, 0, len(artifactPackageList.Packages))
	for _, re := range artifactPackageList.Packages {
		if len(re.LogoImageId) > 0 {
			re.Icon = config.Env.ArtifactHubPackageLogoUrl + re.LogoImageId
		}
		packages = append(packages, re)
	}

	return common.ListRespOK(c, listCount, packages)
}

// GetHelmPackageInfo
// @Summary Get Helm Package Chart Details
// @Tags ArtifactHub
// @Accept json
// @Produce json
// @Router /api/hub/packages/:repositories/:packages [Get]
func GetHelmPackageInfo(c *fiber.Ctx) error {
	repoName := c.Params("repositories")
	packageName := c.Params("packages")

	packageDetailUrl := strings.ReplaceAll(config.Env.ArtifactHubPackageDetail, "{repoName}", repoName)
	packageDetailUrl = strings.ReplaceAll(packageDetailUrl, "{packageName}", packageName)

	reqUrl := fmt.Sprintf("%v%v", config.Env.ArtifactHubUrl, packageDetailUrl)
	respData, err := getRequestData(reqUrl, false)
	if err != nil {
		return common.RespErr(c, err)
	}

	var artifactPackage artifactPackage
	if err := json.Unmarshal(respData.Data, &artifactPackage); err != nil {
		return common.RespErr(c, err)
	}

	if len(artifactPackage.LogoImageId) > 0 {
		artifactPackage.Icon = config.Env.ArtifactHubPackageLogoUrl + artifactPackage.LogoImageId
	}

	return common.RespOK(c, artifactPackage)
}

// GetHelmPackageValues
// @Summary Get Helm Package Chart Values
// @Tags ArtifactHub
// @Accept json
// @Produce json
// @Router /api/hub/packages/:packageID/:version/values [Get]
func GetHelmPackageValues(c *fiber.Ctx) error {
	packageID := c.Params("packageID")
	version := c.Params("version")

	packageValueUrl := strings.ReplaceAll(config.Env.ArtifactHubPackageValues, "{packageID}", packageID)
	packageValueUrl = strings.ReplaceAll(packageValueUrl, "{version}", version)

	reqUrl := fmt.Sprintf("%v%v", config.Env.ArtifactHubUrl, packageValueUrl)
	respData, err := getRequestData(reqUrl, false)
	if err != nil {
		return common.RespErr(c, err)
	}

	return common.RespOK(c, string(respData.Data))
}

func getRequestData(url string, isList bool) (respData, error) {
	log.Infof("SEND :: REQUEST-URL: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		return respData{}, err
	}

	// resp.Body.close()
	defer func() {
		if resp != nil {
			err := resp.Body.Close()
			if err != nil {
				log.Error(err)
			}
		}
	}()

	// if 404 notfound
	if resp.StatusCode == fiber.StatusNotFound {
		return respData{}, fmt.Errorf(common.NOT_FOUND)
	}

	// read body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return respData{}, err
	}

	if !isList {
		return respData{0, data}, nil
	}

	//Pagination-Total-Count
	totalCount, err := strconv.Atoi(resp.Header.Get("Pagination-Total-Count"))
	if err != nil {
		return respData{}, err
	}

	return respData{totalCount, data}, nil
}
