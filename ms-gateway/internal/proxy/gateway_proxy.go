package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/CakeForKit/rsoi-lab2/ms-gateway/internal/util"
)

func NewGatewayProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director:       reverseProxyDirector,       // хук на запрос (до отправки)
		ModifyResponse: reverseProxyModifyResponse, // хук на ответ (до возврата клиенту)
	}
}

func reverseProxyDirector(request *http.Request) {
	route, _ := util.FindRouteByPath(request.URL.Path)
	if err := util.ApplyRequestUri(request, route); err != nil {
		panic(err)
	}
	request.Header.Del("Accept-Encoding")
}

func reverseProxyModifyResponse(response *http.Response) error {
	response.Header.Del("www-authenticate")
	response.Header.Del("Access-Control-Allow-Origin")
	response.Header.Del("Access-Control-Allow-Credentials")
	response.Header.Del("Access-Control-Allow-Methods")
	response.Header.Del("Access-Control-Allow-Headers")
	return nil
}
