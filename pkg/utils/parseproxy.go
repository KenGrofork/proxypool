package utils

import (
	"encoding/json"
	"errors"
	"github.com/vrichv/proxypool/pkg/proxy"
	"github.com/metacubex/mihomo/adapter"
	C "github.com/metacubex/mihomo/constant"
)

func ParseCustomizeProxy(p proxy.Proxy) (cProxy C.Proxy, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return
	}
	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
	}
	if proxy.GoodNodeThatClashUnsupported(p) {
		err = errors.New("not support")
		return
	}
	cProxy, err = adapter.ParseProxy(pmap)
	return
}
