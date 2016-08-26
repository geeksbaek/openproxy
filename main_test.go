package openproxy

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	pf := NewProxyFactory()
	pf.AddDefaultOpenProxySources()
	src := pf.ProxySource()
	for i := 0; i < 100; i++ {
		fmt.Println(<-src)
	}
}

func ExampleProxyFactory_ProxySource() {
	pf := NewProxyFactory()
	pf.AddDefaultOpenProxySources()
	for _ = range pf.ProxySource() {
		// ...
	}
}
