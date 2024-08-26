package stream

import (
	"fmt"
	PC "github.com/vrichv/proxypool/config"
	"github.com/vrichv/proxypool/internal/cache"
	"github.com/vrichv/proxypool/pkg/proxy"
	"github.com/vrichv/proxypool/pkg/utils"
	C "github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/log"
	"time"
)

func RunNetflix(proxies proxy.ProxyList) proxy.ProxyList {
	proxiesList := formatProxiesList(proxies)
	connNum := PC.Config.StreamMaxConn
	//同时连接数
	if i := proxies.Len(); i < connNum {
		connNum = i
	}
	start := time.Now()
	netflixList := BatchCheck(proxiesList, connNum)
	if len(netflixList) == 0 {
		log.Warnln("No unlock node were found.")
		return proxies
	}
	cache.NetflixCount = len(netflixList)
	report := fmt.Sprintf("Total %d nodes test completed, %d unlock nodes, Elapsed time: %s", proxies.Len(), len(netflixList), time.Since(start).Round(time.Millisecond))
	log.Infoln(report)
	return NETFLIXFilter(netflixList, proxies)
}

func RunDisney(proxies proxy.ProxyList) proxy.ProxyList {
	proxiesList := formatProxiesList(proxies)
	start := time.Now()
	connNum := PC.Config.StreamMaxConn
	//同时连接数
	if i := proxies.Len(); i < connNum {
		connNum = i
	}
	disneyList := BatchDisneyCheck(proxiesList, connNum)
	if len(disneyList) == 0 {
		log.Warnln("No unlock node were found.")
		return proxies
	}
	cache.DisneyCount = len(disneyList)
	report := fmt.Sprintf("Total %d nodes test completed, %d unlock nodes, Elapsed time: %s", len(proxiesList), len(disneyList), time.Since(start).Round(time.Millisecond))
	log.Warnln(report)
	return DISNEYFilter(disneyList, proxies)
}

func formatProxiesList(proxies proxy.ProxyList) (proxiesList []C.Proxy) {
	for _, p := range proxies {
		pp, err := utils.ParseCustomizeProxy(p)
		if err != nil {
			continue
		}
		proxiesList = append(proxiesList, pp)
	}
	return proxiesList
}
