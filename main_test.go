package openproxy

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	ps := NewProxySource()
	ps.AddDefaultOpenProxySources()
	src := ps.ProxySource()
	for i := 0; i < 100; i++ {
		fmt.Println(<-src)
	}
}
