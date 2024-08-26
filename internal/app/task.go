package app

import (
	"fmt"
	"github.com/vrichv/proxypool/pkg/geoIp"
	"github.com/vrichv/proxypool/pkg/healthcheck/stream"
	"os"
	"sync"
	"time"

	C "github.com/vrichv/proxypool/config"
	"github.com/vrichv/proxypool/internal/cache"
	"github.com/vrichv/proxypool/internal/database"
	"github.com/vrichv/proxypool/log"
	"github.com/vrichv/proxypool/pkg/healthcheck"
	"github.com/vrichv/proxypool/pkg/provider"
	"github.com/vrichv/proxypool/pkg/proxy"
)

//var location, _ = time.LoadLocation("Asia/Shanghai")

func CrawlGo() {
	wg := &sync.WaitGroup{}
	var pc = make(chan proxy.Proxy)
	for _, g := range Getters {
		wg.Add(1)
		go g.Get2ChanWG(pc, wg)
	}
	proxies := cache.GetProxies("allproxies")
	dbProxies := database.GetAllProxies()
	// Show last time result when launch
	if proxies == nil && dbProxies != nil {
		cache.SetProxies("proxies", dbProxies)
		cache.LastCrawlTime = "抓取中，已载入上次数据库数据"
		log.Infoln("Database: loaded")
	}
	if dbProxies != nil {
		proxies = dbProxies.UniqAppendProxyList(proxies)
	}
	if proxies == nil {
		proxies = make(proxy.ProxyList, 0)
	}

	go func() {
		wg.Wait()
		close(pc)
	}() // Note: 为何并发？可以一边抓取一边读取而非抓完再读
	// for 用于阻塞goroutine
	for p := range pc { // Note: pc关闭后不能发送数据可以读取剩余数据
		if p != nil {
			proxies = proxies.UniqAppendProxy(p)
		}
	}

	proxies.NameClear()
	proxies = proxies.Derive()
	if C.Config.OnlyNode {
		clash := provider.Clash{
			Base: provider.Base{
				Proxies: &proxies,
			},
		}
		text := clash.Provide()

		file, err := os.OpenFile("allnode.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Infoln("Error opening file")
		}
		defer file.Close()
		_, err = file.WriteString(text)
		if err != nil {
			log.Infoln("Error writing to file")
		}
		log.Infoln("String successfully written to file.")
	}
	log.Infoln("CrawlGo unique proxy count: %d", len(proxies))

	// Clean Clash unsupported proxy because health check depends on clash
	proxies = provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.CleanProxies()
	log.Infoln("CrawlGo clash supported proxy count: %d", len(proxies))

	cache.SetProxies("allproxies", proxies)
	cache.AllProxiesCount = proxies.Len()
	log.Infoln("AllProxiesCount: %d", cache.AllProxiesCount)
	cache.SSProxiesCount = proxies.TypeLen("ss")
	log.Infoln("SSProxiesCount: %d", cache.SSProxiesCount)
	cache.SSRProxiesCount = proxies.TypeLen("ssr")
	log.Infoln("SSRProxiesCount: %d", cache.SSRProxiesCount)
	cache.VmessProxiesCount = proxies.TypeLen("vmess")
	log.Infoln("VmessProxiesCount: %d", cache.VmessProxiesCount)
	cache.TrojanProxiesCount = proxies.TypeLen("trojan")
	log.Infoln("TrojanProxiesCount: %d", cache.TrojanProxiesCount)
	cache.LastCrawlTime = time.Now().In(time.FixedZone("UTC+8", 8*60*60)).Format("2006-01-02 15:04:05")
	cache.Hysteria2ProxiesCount = proxies.TypeLen("hysteria2")
	log.Infoln("Hysteria2ProxiesCount: %d", cache.Hysteria2ProxiesCount)
	cache.HysteriaProxiesCount = proxies.TypeLen("hysteria")
	log.Infoln("HysteriaProxiesCount: %d", cache.HysteriaProxiesCount)
	cache.VlessProxiesCount = proxies.TypeLen("vless")
	log.Infoln("VlessProxiesCount: %d", cache.VlessProxiesCount)
	cache.TuicProxiesCount = proxies.TypeLen("tuic")
	log.Infoln("TuicProxiesCount: %d", cache.TuicProxiesCount)
	// Health Check
	log.Infoln("Now proceed proxy health check...")
	healthcheck.SpeedConn = C.Config.SpeedConnection
	healthcheck.DelayConn = C.Config.HealthCheckConnection
	if C.Config.HealthCheckTimeout > 0 {
		healthcheck.DelayTimeout = time.Second * time.Duration(C.Config.HealthCheckTimeout)
		log.Infoln("CONF: Health check timeout is set to %d seconds", C.Config.HealthCheckTimeout)
	}
	proxies = healthcheck.CleanBadProxiesWithGrpool(proxies)
	// proxies = healthcheck.CleanBadProxies(proxies)
	log.Infoln("CrawlGo clash usable proxy count: %d", len(proxies))
	// Format name like US_01 sorted by country
	proxies.NameAddCounrty().Sort()
	log.Infoln("Proxy rename DONE!")
	// Relay check and rename
	proxies = healthcheck.RelayCheck(proxies)
	for i := range proxies {
		if s, ok := healthcheck.ProxyStats.Find(proxies[i]); ok {
			log.Debugln("ip: %s ,outip: %s,delay:%s",proxies[i].BaseInfo().Server,s.OutIp,s.Delay)
			_, c, o, e := geoIp.GeoIpDB.Find(s.OutIp, true)
			if e == nil {
				nameStr := proxies[i].BaseInfo().Name
				if proxies[i].BaseInfo().Name == "" {
					nameStr = proxies[i].BaseInfo().Country
				}
				if s.Relay {
						proxies[i].SetName(fmt.Sprintf("Relay_%s-%s[%s]", nameStr, c, o))					
				} else if s.Pool {
						proxies[i].SetName(fmt.Sprintf("Pool_%s[%s]", nameStr, o))
				} else {
						proxies[i].SetName(fmt.Sprintf("%s[%s]", nameStr, o))
				}
			}
		}
	}
	proxies.NameAddIndex()
	if C.Config.NetflixTest {
		cache.IsNetflixTest = "已开启"
		proxies = stream.RunNetflix(proxies)
		log.Infoln("Netflix check DONE!")
	} else {
		cache.IsNetflixTest = "未开启"
	}
	if C.Config.DisneyTest {
		cache.IsDisneyTest = "已开启"
		stream.RunDisney(proxies)
		log.Infoln("Disney check DONE!")
	} else {
		cache.IsDisneyTest = "未开启"
	}
	// 可用节点存储
	cache.SetProxies("proxies", proxies)
	cache.UsefullProxiesCount = proxies.Len()
	database.SaveProxyList(proxies)
	database.ClearOldItems()
	log.Infoln("Usablility checking done. Open %s to check", C.Config.HostUrl())
	if C.Config.OnlyNode {
		clash := provider.Clash{
			Base: provider.Base{
				Proxies: &proxies,
			},
		}
		text := clash.Provide()

		file, err := os.OpenFile("output.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Infoln("Error opening file")
		}
		defer file.Close()
		_, err = file.WriteString(text)
		if err != nil {
			log.Infoln("Error writing to file")
		}
		log.Infoln("String successfully written to file.")
		os.Exit(0)
	}
	// 测速
	speedTestNew(proxies)
	cache.SetString("clashproxies", provider.Clash{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide()) // update static string provider
	cache.SetString("surgeproxies", provider.Surge{
		Base: provider.Base{
			Proxies: &proxies,
		},
	}.Provide())
}

// Speed test for new proxies
func speedTestNew(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestNew(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}

// Speed test for all proxies in proxy.ProxyList
func SpeedTest(proxies proxy.ProxyList) {
	if C.Config.SpeedTest {
		cache.IsSpeedTest = "已开启"
		if C.Config.SpeedTimeout > 0 {
			log.Infoln("config: Speed test timeout is set to %d seconds", C.Config.SpeedTimeout)
			healthcheck.SpeedTimeout = time.Second * time.Duration(C.Config.SpeedTimeout)
		}
		healthcheck.SpeedTestAll(proxies)
	} else {
		cache.IsSpeedTest = "未开启"
	}
}
