package document

import (
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/hugo19941994/helm-docs/pkg/helm"
)

type DependencyValues struct {
	Prefix                  string
	ChartValues             *yaml.Node
	ChartValuesDescriptions map[string]helm.ChartValueDescription
}

func GetDependencyValues(root helm.ChartDocumentationInfo, allChartInfoByChartPath map[string]helm.ChartDocumentationInfo) ([]DependencyValues, error) {
	return getDependencyValuesWithPrefix(root, allChartInfoByChartPath, "")
}

func getDependencyValuesWithPrefix(root helm.ChartDocumentationInfo, allChartInfoByChartPath map[string]helm.ChartDocumentationInfo, prefix string) ([]DependencyValues, error) {
	if len(root.Dependencies) == 0 {
		return nil, nil
	}

	result := make([]DependencyValues, 0, len(root.Dependencies))

	for _, dep := range root.Dependencies {
		searchPath := ""

		if strings.HasPrefix(dep.Repository, "file://") {
			searchPath = filepath.Join(root.ChartDirectory, strings.TrimPrefix(dep.Repository, "file://"))
		} else if strings.HasPrefix(dep.Repository, "oci://") {
			searchPath = filepath.Join(root.ChartDirectory, "charts", dep.Name)
		} else {
			searchPath = filepath.Join(root.ChartDirectory, "charts", dep.Name)
		}

		depInfo, ok := allChartInfoByChartPath[searchPath]
		if !ok {
			log.Warnf("Dependency with path %q was not found. Dependency values will not be included.", searchPath)
			continue
		}

		alias := dep.Alias
		if alias == "" {
			alias = dep.Name
		}
		depPrefix := prefix + alias

		result = append(result, DependencyValues{
			Prefix:                  depPrefix,
			ChartValues:             depInfo.ChartValues,
			ChartValuesDescriptions: depInfo.ChartValuesDescriptions,
		})

		children, err := getDependencyValuesWithPrefix(depInfo, allChartInfoByChartPath, depPrefix+".")
		if err != nil {
			return nil, err
		}

		result = append(result, children...)
	}

	return result, nil
}
