package openproxy

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// ProxyBuilder 구조체는 프록시를 가져올 Source와 Parser 함수로
// 이루어져 있습니다. Source는 여러 개일 수 있으나, 단일한 Parser 함수로
// 호환 가능한 형태여야 합니다. Parser는 func(source string) []*url.URL 형태를
// 충족시켜야 합니다.
type ProxyBuilder struct {
	Source []string
	Parser func(source string) []*url.URL
}

// ProxyFactory 구조체는 복수의 ProxySource를 가집니다.
type ProxyFactory struct {
	ProxyBuilders []*ProxyBuilder
}

var defaultProxySources = []*ProxyBuilder{{
	Source: []string{
		"https://incloak.com/proxy-list/?type=hs#list",
		"https://incloak.com/proxy-list/?type=hs&start=64#list",
		"https://incloak.com/proxy-list/?type=hs&start=128#list",
		"https://incloak.com/proxy-list/?type=hs&start=192#list",
		"https://incloak.com/proxy-list/?type=hs&start=256#list",
		"https://incloak.com/proxy-list/?type=hs&start=320#list",
		"https://incloak.com/proxy-list/?type=hs&start=384#list",
		"https://incloak.com/proxy-list/?type=hs&start=448#list",
		"https://incloak.com/proxy-list/?type=hs&start=512#list",
		"https://incloak.com/proxy-list/?type=hs&start=576#list",
		"https://incloak.com/proxy-list/?type=hs&start=640#list",
	},
	Parser: func(URL string) (proxys []*url.URL) {
		doc, err := goquery.NewDocument(URL)
		if err != nil {
			return
		}
		doc.Find(`#content-section > section.proxy > div > table > tbody > tr`).Each(func(i int, s *goquery.Selection) {
			ip := s.Find(`.tdl`).Text()
			port := s.Find(`td:nth-child(2)`).Text()
			proxy, err := url.Parse(fmt.Sprintf("http://%s:%s", ip, port))
			if err == nil {
				proxys = append(proxys, proxy)
			}
		})
		return
	}}, {
	Source: []string{
		"https://www.us-proxy.org/",
		"https://free-proxy-list.net/",
	},
	Parser: func(URL string) (proxys []*url.URL) {
		doc, err := goquery.NewDocument(URL)
		if err != nil {
			return
		}
		doc.Find(`#proxylisttable > tbody > tr`).Each(func(i int, s *goquery.Selection) {
			ip := s.Find(`td:nth-child(1)`).Text()
			port := s.Find(`td:nth-child(2)`).Text()
			proxy, err := url.Parse(fmt.Sprintf("http://%s:%s", ip, port))
			if err == nil {
				proxys = append(proxys, proxy)
			}
		})
		return
	},
}, {
	Source: []string{"https://nordvpn.com/wp-admin/admin-ajax.php?searchParameters%5B2%5D%5Bname%5D=http&searchParameters%5B2%5D%5Bvalue%5D=on&searchParameters%5B3%5D%5Bname%5D=https&searchParameters%5B3%5D%5Bvalue%5D=on&offset=0&limit=10000&action=getProxies"},
	Parser: func(URL string) (proxys []*url.URL) {
		resp, err := http.Get(URL)
		if err != nil {
			return
		}
		var jsonProxys []struct {
			IP   string
			Port string
			Type string
		}
		json.NewDecoder(resp.Body).Decode(&jsonProxys)
		for _, jsonProxy := range jsonProxys {
			proxy, err := url.Parse(fmt.Sprintf("%s://%s:%s", jsonProxy.Type, jsonProxy.IP, jsonProxy.Port))
			if err == nil {
				proxys = append(proxys, proxy)
			}
		}
		return
	},
}}

// NewProxyFactory 함수는 빈 ProxyFactory 구조체를 반환합니다.
func NewProxyFactory() *ProxyFactory {
	return &ProxyFactory{}
}

// AddDefaultOpenProxySources 함수는 패키지에 포함되어 있는
// 기본 ProxySource들을 추가합니다.
func (pf *ProxyFactory) AddDefaultOpenProxySources() {
	for _, defaultProxySource := range defaultProxySources {
		pf.ProxyBuilders = append(pf.ProxyBuilders, defaultProxySource)
	}
}

// AddCustomProxySource 함수는 사용자가 구현한 ProxyBuilder를
// 인자로 받아 ProxyFactory 객체에 추가합니다.
func (pf *ProxyFactory) AddCustomProxySource(proxyBuilder *ProxyBuilder) {
	pf.ProxyBuilders = append(pf.ProxyBuilders, proxyBuilder)
}

// ProxySource 함수는 프록시를 전송하는 채널을 반환합니다. 이 채널은 파싱된
// 프록시 url을 무작위로 계속해서 전송하며 채널은 계속 열려있습니다.
func (pf *ProxyFactory) ProxySource() chan *url.URL {
	retCh := make(chan *url.URL)
	bufCh := make(chan *url.URL)

	go func() {
		wg := new(sync.WaitGroup)
		for _, ProxyBuilder := range pf.ProxyBuilders {
			for _, source := range ProxyBuilder.Source {
				wg.Add(1)
				go func(source string) {
					defer wg.Done()
					for _, proxy := range ProxyBuilder.Parser(source) {
						bufCh <- proxy
					}
				}(source)
			}
		}
		wg.Wait()
		close(bufCh)
	}()

	go func() {
		proxys := []*url.URL{}
		for proxy := range bufCh {
			proxys = append(proxys, proxy)
		}
		shuffle(proxys)
		for i := 0; ; i++ {
			if i == len(proxys) {
				i = 0
			}
			retCh <- proxys[i]
		}
	}()
	return retCh
}

func shuffle(a []*url.URL) {
	rand.Seed(time.Now().UTC().UnixNano())
	for i := range a {
		j := rand.Intn(i + 1)
		a[i], a[j] = a[j], a[i]
	}
}
