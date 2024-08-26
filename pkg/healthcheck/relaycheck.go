package healthcheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/vrichv/proxypool/log"
	"github.com/vrichv/proxypool/pkg/proxy"
	"github.com/ivpusic/grpool"
	"github.com/metacubex/mihomo/adapter"
)

func RelayCheck(proxies []proxy.Proxy)(cproxies []proxy.Proxy)  {
	numWorker := DelayConn
	numJob := 1
	if numWorker > 4 {
		numJob = (numWorker + 2) / 3
	}

	pool := grpool.NewPool(numWorker, numJob)
	pool.WaitCount(len(proxies))
	cproxies = make(proxy.ProxyList, 0, 500)
	m := sync.Mutex{}

	log.Infoln("Relay Test ON")
	//doneCount := 0
	//dcm := sync.Mutex{}
	go func() {
		for _, p := range proxies {
			pp := p
			pool.JobQueue <- func() {
				defer pool.JobDone()
				out, err := testRelay(pp)
				m.Lock()
				if err != nil {					
					//测试出错，判定节点异常
					if ps, ok := ProxyStats.Find(pp); ok {
						ps.UpdatePSDelay(0)
					}				
				}else {
					cproxies = append(cproxies, pp)
					if out == "" {
						//not relay
						if ps, ok := ProxyStats.Find(pp); ok {
							ps.UpdatePSOutIp(pp.BaseInfo().Server)
						} else {
							ps = &Stat{
								Id:    pp.Identifier(),
								OutIp: pp.BaseInfo().Server,
							}
							ProxyStats = append(ProxyStats, *ps)
						}
					}else{
						// Relay or pool
						if isRelay(pp.BaseInfo().Server, out) {
							if ps, ok := ProxyStats.Find(pp); ok {
								ps.UpdatePSOutIp(out)
								ps.Relay = true
							} else {
								ps = &Stat{
									Id:    pp.Identifier(),
									Relay: true,
									OutIp: out,
								}
								ProxyStats = append(ProxyStats, *ps)
							}
						} else { // is pool ip
							if ps, ok := ProxyStats.Find(pp); ok {
								ps.UpdatePSOutIp(out)
								ps.Pool = true
							} else {
								ps = &Stat{
									Id:    pp.Identifier(),
									Pool:  true,
									OutIp: out,
								}
								ProxyStats = append(ProxyStats, *ps)
							}
						}
						
					} 					
				}
				m.Unlock()
				// dcm.Lock()
				// doneCount++
				// progress := (doneCount * 100) / len(proxies)
				// if progress%20 == 0 && progress > 0 { 
				// 	fmt.Printf("\r\t[%d%% DONE]", progress)
				// }
				// dcm.Unlock()
			}
		}
	}()
	pool.WaitAll()
	pool.Release()
	fmt.Println()
	return
}

// Get outbound relay ip
func testRelay(p proxy.Proxy) (outip string, err error) {
	pmap := make(map[string]interface{})
	err = json.Unmarshal([]byte(p.String()), &pmap)
	if err != nil {
		return "", err
	}

	pmap["port"] = int(pmap["port"].(float64))
	if p.TypeName() == "vmess" {
		pmap["alterId"] = int(pmap["alterId"].(float64))
		if network, ok := pmap["network"]; ok && network.(string) == "h2" {
			return "", nil // todo 暂无方法测试h2的延迟，clash对于h2的connection会阻塞
		}
	}

	if proxy.GoodNodeThatClashUnsupported(p) {
		host := pmap["server"].(string)
		port := fmt.Sprint(pmap["port"].(int))
		if result, _, err := netConnectivity(host, port); err == nil {
			return result, nil
		} else {
			return "", err
		}
	}

	clashProxy, err := adapter.ParseProxy(pmap)
	if err != nil {
		return "", err
	}

	b, err := HTTPGetBodyViaProxyWithTime(clashProxy, "http://ipinfo.io/ip", RelayTimeout)
	if err != nil {
		return "", err
	}

	if string(b) == p.BaseInfo().Server {
		return "", nil // not relay
	}

	address := net.ParseIP(string(b))
	if address == nil {
		return "", errors.New("error outbound ip format")
	}

	return string(b), nil
}

// Distinguish pool ip from relay. false for pool, true for relay
func isRelay(src string, out string) bool {
	ipv4Mask := net.CIDRMask(16, 32)
	ip1 := net.ParseIP(src)
	ip2 := net.ParseIP(out)
	return fmt.Sprint(ip1.Mask(ipv4Mask)) != fmt.Sprint(ip2.Mask(ipv4Mask))
}
