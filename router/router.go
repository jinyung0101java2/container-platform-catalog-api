package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"go-api/common"
	"go-api/handler"
)

func AccessibleRoute(app *fiber.App) {
	// swagger
	route := app.Group("/swagger")
	{
		route.Get("*", swagger.HandlerDefault)
	}

	// health check
	health := app.Group("/actuator/health")
	{
		health.Get("", handler.Health)
		health.Get("/liveness", handler.HealthCheck)
		health.Get("/readiness", handler.HealthCheck)
	}
}
func APIRoutes(app *fiber.App) {
	api := app.Group("/api")

	// repositories
	repositories := api.Group("/repositories")
	{
		// helm repo list
		repositories.Get("", handler.ListRepos)
		// helm repo add
		repositories.Post("", handler.AddRepo)
		// helm repo remove
		repositories.Delete("/:repositories", handler.RemoveRepo)
		// helm repo update
		repositories.Put("/:repositories", handler.UpdateRepo)
		// helm search chart list
		repositories.Get("/:repositories/charts", handler.ListRepoCharts)
		// helm show chart
		repositories.Get("/:repositories/charts/:charts/info", handler.GetChartInfo)
		// clear cache
		repositories.Delete("/cache/clear", handler.ClearRepoCache)
	}

	// artifactHub
	artifact := api.Group("/hub")
	{
		// artifactHub search repo
		artifact.Get("/repositories", handler.SearchRepoHub)
		// artifactHub search package
		artifact.Get("/packages", handler.SearchPackageHub)
		// artifactHub get package details
		artifact.Get("/packages/:repositories/:packages", handler.GetHelmPackageInfo)
		// artifactHub get package values
		artifact.Get("/packages/:packageID/:version/values", handler.GetHelmPackageValues)
	}

	// releases
	releases := api.Group("/clusters/:clusterId/namespaces/:namespace/releases")
	{
		// helm list
		releases.Get("", handler.ListReleases).Name(common.LIST_RELEASES)
		// helm get
		releases.Get("/:release", handler.GetReleaseInfo)
		//helm install
		releases.Post("/:release", handler.InstallRelease)
		//helm upgrade
		releases.Put("/:release", handler.UpgradeRelease)
		// helm rollback
		releases.Put("/:release/versions/:revision", handler.RollbackRelease)
		// helm uninstall
		releases.Delete("/:release", handler.UninstallRelease)
		// helm release history
		releases.Get("/:release/histories", handler.GetReleaseHistories)
		// helm release resources status
		releases.Get("/:release/resources", handler.GetReleaseResources)
	}

	charts := api.Group("/charts/:charts")
	{ // helm search charts (all versions)
		charts.Get("/versions", handler.GetChartVersions)
	}

}
