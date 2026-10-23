package util

import (
	"net/http"
	"strings"

	coreError "github.com/CakeForKit/rsoi-lab2/lb-core/custom_error"
	"github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/config"
	"github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/model"
)

func FindRouteByPath(path string) (*model.Route, error) {
	var result *model.Route
	for _, propsRoute := range config.AppConfig.Application.Gateway.Routes {
		if strings.HasPrefix(path, propsRoute.PathPrefix) {
			result = &propsRoute
			break
		}
	}
	return result, nil
}

func ApplyRequestUri(request *http.Request, route *model.Route) error {
	schemeAndHost := strings.Split(route.Uri, "://")
	if len(schemeAndHost) != 2 {
		return coreError.InvalidUriError(route.Uri)
	}
	scheme, host := schemeAndHost[0], schemeAndHost[1]

	(*request).URL.Scheme = scheme
	(*request).Host, (*request).URL.Host = host, host
	(*request).RequestURI = ""
	(*request).URL.Path = strings.TrimPrefix(request.URL.Path, route.PathPrefix)

	return nil
}
