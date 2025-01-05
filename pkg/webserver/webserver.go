package webserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/webbgeorge/castkeeper/pkg/components/pages"
	"github.com/webbgeorge/castkeeper/pkg/framework"
	"github.com/webbgeorge/castkeeper/pkg/podcasts"
	"github.com/webbgeorge/castkeeper/web"
	"go.opentelemetry.io/otel/sdk/resource"
	"gorm.io/gorm"
)

func Start(ctx context.Context, otelRes *resource.Resource, db *gorm.DB) error {
	server, err := framework.NewServer(otelRes, ":8080")
	if err != nil {
		return fmt.Errorf("failed to start server", err)
	}

	middleware := framework.DefaultMiddlewareStack()

	return server.SetServerMiddlewares(middleware...).
		AddFileServer("GET /static/", http.FileServer(http.FS(web.StaticAssets))).
		AddRoute("GET /", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if r.URL.Path != "/" {
				return framework.HttpNotFound()
			}
			return framework.Render(ctx, w, 200, pages.Home())
		}).
		AddRoute("GET /podcasts/subscribe", podcasts.NewSubscribeGetHandler()).
		AddRoute("POST /podcasts/subscribe", podcasts.NewSubscribePostHandler(db)).
		Start(ctx)
}
