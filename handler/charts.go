package handler

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"go-api/common"
	"helm.sh/helm/v3/cmd/helm/search"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/helmpath"
	"helm.sh/helm/v3/pkg/repo"
	"path/filepath"
	"strings"
)

// searchMaxScore suggests that any score higher than this is not considered a match.
const searchMaxScore = 25

var readmeFileNames = []string{"readme.md", "readme.txt", "readme"}

type file struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type repoChartElement struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"app_version"`
	Home        string `json:"home"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	RepoName    string `json:"repoName"`
	Deprecated  bool   `json:"deprecated"`
}

type repoChartList []repoChartElement

// GetChartVersions
// @Summary Get Chart Versions
// @Tags Charts
// @Accept json
// @Produce json
// @Router /api/charts/:charts/versions [Get]
func GetChartVersions(c *fiber.Ctx) error {
	charts := c.Params("charts")    // search keyword
	repoName := c.Query("repo", "") // repo name
	version := c.Query("version")
	// default stable
	if version == "" {
		version = ">0.0.0"
	}

	log.Infof("GetChartVersions:: repoName: %v, charts: %v", repoName, charts)
	var index *search.Index
	var err error
	var keyword string

	if len(repoName) < 1 {
		// search in all repos
		index, err = buildSearchIndexAll()
		keyword = fmt.Sprintf("/%s\v", charts)
	} else {
		index, err = buildSearchIndex(repoName)
		keyword = fmt.Sprintf("\v%s/%s\v", repoName, charts)
	}

	if err != nil {
		return common.RespErr(c, err)
	}

	var res []*search.Result
	res, err = index.Search(keyword, searchMaxScore, true)
	if err != nil {
		return common.RespErr(c, err)
	}

	search.SortScore(res)
	data, err := applyConstraint(version, true, res)
	if err != nil {
		return common.RespErr(c, err)
	}

	if len(data) < 1 {
		return common.RespErr(c, fmt.Errorf(common.CHART_NOT_FOUND))
	}

	chartList := make(repoChartList, 0, len(data))
	for _, v := range data {
		chartList = append(chartList, repoChartElement{
			Name:       v.Name,
			Version:    v.Chart.Version,
			AppVersion: v.Chart.AppVersion,
			Home:       v.Chart.Home,
			Deprecated: v.Chart.Deprecated,
		})
	}

	return common.RespOK(c, chartList)
}

// GetChartInfo
// @Summary Get Chart Info
// @Tags Repository
// @Accept json
// @Produce json
// @Router /api/repositories/:repositories/charts/:charts/info [Get]
func GetChartInfo(c *fiber.Ctx) error {
	repoName := c.Params("repositories")
	charts := c.Params("charts") // search keyword
	version := c.Query("version")
	info := c.Query("info") // all, readme, values, chart

	if info == "" {
		info = string(action.ShowAll)
	}

	actionConfig := new(action.Configuration)
	client := action.NewShowWithConfig(action.ShowAll, actionConfig)
	client.Version = version
	if info == string(action.ShowChart) {
		client.OutputFormat = action.ShowChart
	} else if info == string(action.ShowReadme) {
		client.OutputFormat = action.ShowReadme
	} else if info == string(action.ShowValues) {
		client.OutputFormat = action.ShowValues
	} else if info == string(action.ShowAll) {
		client.OutputFormat = action.ShowAll
	} else {
		return common.RespErr(c, fmt.Errorf("chart info only support readme/values/chart"))
	}

	aimChart := fmt.Sprintf("%s/%s", repoName, charts)
	cp, err := client.ChartPathOptions.LocateChart(aimChart, settings)
	if err != nil {
		return common.RespErr(c, err)
	}

	chrt, err := loader.Load(cp)
	if err != nil {
		return common.RespErr(c, err)
	}

	if client.OutputFormat == action.ShowChart {
		return common.RespOK(c, chrt.Metadata)
	}
	if client.OutputFormat == action.ShowValues {
		var values string
		for _, v := range chrt.Raw {
			if v.Name == chartutil.ValuesfileName {
				values = string(v.Data)
				break
			}
		}
		return common.RespOK(c, values)
	}
	if client.OutputFormat == action.ShowReadme {
		return common.RespOK(c, string(findReadme(chrt.Files).Data))
	}
	if client.OutputFormat == action.ShowAll {
		values := make([]*file, 0, len(chrt.Raw))
		for _, v := range chrt.Raw {
			values = append(values, &file{
				Name: v.Name,
				Data: string(v.Data),
			})
		}

		return common.RespOK(c, values)
	}
	return common.RespOK(c, nil)
}

func findReadme(files []*chart.File) (file *chart.File) {
	for _, file := range files {
		for _, n := range readmeFileNames {
			if file == nil {
				continue
			}
			if strings.EqualFold(file.Name, n) {
				return file
			}
		}
	}
	return nil
}

func buildSearchIndex(repoName string) (*search.Index, error) {
	index := search.NewIndex()
	path := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(repoName))
	indexFile, err := repo.LoadIndexFile(path)
	if err != nil {
		return nil, fmt.Errorf(common.REPO_CORRUPT_MISSING)
	}

	index.AddRepo(repoName, indexFile, true)
	return index, nil
}

func buildSearchIndexAll() (*search.Index, error) {
	repos, err := repo.LoadFile(settings.RepositoryConfig)
	switch {
	case isNotExist(err):
		return nil, fmt.Errorf(common.REPO_NO_CONFIGURED)
	case err != nil:
		return nil, fmt.Errorf(common.REPO_FAILED_LOADING_FILE)
	case len(repos.Repositories) == 0:
		return nil, fmt.Errorf(common.REPO_NO_CONFIGURED)
	}

	index := search.NewIndex()

	for _, re := range repos.Repositories {
		repoName := re.Name
		f := filepath.Join(settings.RepositoryCache, helmpath.CacheIndexFile(repoName))
		indexFile, err := repo.LoadIndexFile(f)
		if err != nil {
			continue
		}
		index.AddRepo(repoName, indexFile, true)
	}
	return index, nil
}

func applyConstraint(version string, versions bool, res []*search.Result) ([]*search.Result, error) {
	if len(version) == 0 {
		return res, nil
	}

	constraint, err := semver.NewConstraint(version)
	if err != nil {
		return res, fmt.Errorf(common.CHART_VERSION_INVALID)
	}

	data := res[:0]
	foundNames := map[string]bool{}
	for _, r := range res {
		// if not returning all versions and already have found a result,
		// you're done!
		if !versions && foundNames[r.Name] {
			continue
		}
		v, err := semver.NewVersion(r.Chart.Version)
		if err != nil {
			continue
		}
		if constraint.Check(v) {
			data = append(data, r)
			foundNames[r.Name] = true
		}
	}

	return data, nil
}
