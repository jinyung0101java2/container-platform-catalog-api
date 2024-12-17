package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"go-api/common"
	"go-api/config"
	"helm.sh/helm/v3/pkg/cli"
	"io"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/testapigroup/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"os"
	"path/filepath"
	"reflect"
	sigyaml "sigs.k8s.io/yaml"
	"strconv"
	"strings"
)

var (
	settings = cli.New()
)

type ListSearchElement struct {
	Offset     int
	Limit      int
	SearchName string
}

func Settings() {
	settings.RepositoryConfig = config.Env.HelmRepoConfig
	settings.RepositoryCache = config.Env.HelmRepoCache
}

func GetResources(out string) []*v1.Carp {
	res, err := ParseManifests(out)
	if err != nil {
		res = append(res, &v1.Carp{
			TypeMeta: metav1.TypeMeta{Kind: "ManifestParseError"},
			ObjectMeta: metav1.ObjectMeta{
				Name: err.Error(),
			},
			Spec: v1.CarpSpec{},
			Status: v1.CarpStatus{
				Phase:   "BrokenManifest",
				Message: err.Error(),
			},
		})
		//_ = c.AbortWithError(http.StatusInternalServerError, err)
		//return
	}
	return res
}

func ParseManifests(out string) ([]*v1.Carp, error) {
	dec := yaml.NewYAMLOrJSONDecoder(strings.NewReader(out), 4096)
	res := make([]*v1.Carp, 0)
	var tmp interface{}
	for {
		err := dec.Decode(&tmp)
		if err == io.EOF {
			break
		}

		if err != nil {
			return res, err
		}

		jsoned, err := json.Marshal(tmp)
		if err != nil {
			return res, err
		}

		var doc v1.Carp
		err = json.Unmarshal(jsoned, &doc)
		if err != nil {
			return res, err
		}

		if doc.Kind == "" {
			log.Warnf("Manifest piece is not k8s resource: %s", jsoned)
			continue
		}

		res = append(res, &doc)
	}
	return res, nil
}

func ConvertYAML(results map[string]interface{}) string {
	obj, err := sigyaml.Marshal(results)
	if err != nil {
		return common.EMPTY_STR
	}
	return string(obj)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func RemoveFile(filename string) error {
	if len(filename) > 0 {
		log.Infof("Delete file :: %s", filename)
		err := os.Remove(filename)
		if err != nil {
			return err
		}

	}
	return nil
}

func ListSearchCheck(c *fiber.Ctx) (*ListSearchElement, error) {
	offset, err := strconv.Atoi(c.Query("offset", "0"))
	if err != nil {
		return nil, fmt.Errorf(common.OFFSET_VAL_INVALID)
	}
	limit, err := strconv.Atoi(c.Query("limit", "0"))
	if err != nil {
		return nil, fmt.Errorf(common.LIMIT_VAL_INVALID)
	}
	if offset < 0 {
		return nil, fmt.Errorf(common.OFFSET_ILLEGAL_ARGUMENT)
	}
	if limit < 0 {
		return nil, fmt.Errorf(common.LIMIT_ILLEGAL_ARGUMENT)
	}
	if offset > 0 && limit == 0 {
		return nil, fmt.Errorf(common.OFFSET_REQUIRES_LIMIT_ILLEGAL_ARGUMENT)
	}

	lse := ListSearchElement{
		Offset:     offset,
		Limit:      limit,
		SearchName: strings.TrimSpace(c.Query("searchName", "")),
	}

	return &lse, nil
}

func ResourceListProcessing(list []interface{}, lse *ListSearchElement) (common.ListCount, []interface{}) {
	// 1. search keyword
	if lse.SearchName != "" {
		list = searchResourceName(list, lse.SearchName)
	}

	// 2. paging (offset & limit)
	allItemCount := len(list)
	if allItemCount < 1 {
		return common.ListCount{}, make([]interface{}, 0)
	}

	remainingItemCount := allItemCount - ((lse.Offset + 1) * lse.Limit)
	start := lse.Offset * lse.Limit

	if lse.Limit == 0 || remainingItemCount < 0 {
		remainingItemCount = 0
	}

	listCount := common.ListCount{
		AllItemCount:       allItemCount,
		RemainingItemCount: remainingItemCount,
	}

	if lse.Limit == 0 {
		return listCount, list
	}
	if start > allItemCount {
		return listCount, make([]interface{}, 0)
	}
	if (start + lse.Limit) > allItemCount {
		return listCount, list[start:]
	}

	return listCount, list[start : start+lse.Limit]
}

func searchResourceName(list []interface{}, searchName string) []interface{} {
	var searchList []interface{}
	for _, re := range list {
		name := reflect.ValueOf(re).FieldByName("Name").String()
		if strings.Contains(name, searchName) {
			searchList = append(searchList, re)
		}
	}
	return searchList
}

func RemoveGlob(path string) (err error) {
	contents, err := filepath.Glob(path)
	if err != nil {
		return
	}
	for _, item := range contents {
		err = os.RemoveAll(item)
		if err != nil {
			log.Infof("Error removing files: %+v", err)
		}
	}
	return
}

func generatingId() string {
	return uuid.New().String()
}

func procReplaceEmpty(value string) string {
	if len(value) < 1 {
		return common.EMPTY_STR
	}
	return value
}
