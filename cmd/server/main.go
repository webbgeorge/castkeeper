package main

import (
	"context"
	"log"
	"net/http"

	"github.com/webbgeorge/castkeeper"
	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/config"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/web"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

const (
	applicationName = "castkeeper"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	otelRes, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(applicationName),
			semconv.DeploymentEnvironmentKey.String(cfg.EnvName),
			semconv.ServiceInstanceIDKey.String(framework.GetHostID()),
			semconv.ServiceVersionKey.String(castkeeper.Version),
		),
	)
	if err != nil {
		log.Fatalf("failed to create otel configuration", err)
	}

	server, err := framework.NewServer(otelRes, ":8080")
	if err != nil {
		log.Fatalf("failed to start server", err)
	}

	middleware := framework.DefaultMiddlewareStack()

	server.SetServerMiddlewares(middleware...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets))).
		AddRoute("GET /", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if r.URL.Path != "/" {
				return framework.HttpNotFound()
			}
			return framework.Render(ctx, w, 200, pages.Home())
		}).
		Start()
}
