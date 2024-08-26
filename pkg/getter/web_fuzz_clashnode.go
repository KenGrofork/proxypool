package getter

import (
	"io"
	"strings"
	"sync"

	"github.com/vrichv/proxypool/log"
	"github.com/vrichv/proxypool/pkg/proxy"

	"github.com/vrichv/proxypool/pkg/tool"
)

func init() {
	Register("webfuzzclashnode", WebFuzzClashNodeGetter)
}

type WebFuzzClashNode struct {
	Url string
}

func (w *WebFuzzClashNode) Get() proxy.ProxyList {

	subUrls := GetUrls(w.Url)
	result := make(proxy.ProxyList, 0)
	if len(subUrls) != 0 {
		for _, url := range subUrls {
			//if (strings.Contains(url, "https://") || strings.Contains(url, "http://")) && strings.HasSuffix(url, "htm") {
			if strings.HasSuffix(url, "htm") {
				wholeUrl := "https://www.freeclashnode.com/free-node/" + url
				pUrls := GetUrls(wholeUrl)
				if len(pUrls) != 0 {
					for _, pUrl := range pUrls {
						if (strings.Contains(pUrl, "https://") || strings.Contains(pUrl, "http://")) && (strings.HasSuffix(pUrl, "txt") || strings.HasSuffix(pUrl, "yaml")) {
							newResult := (&Subscribe{Url: pUrl}).Get()
							if len(newResult) == 0 {
								newResult = (&Clash{Url: pUrl}).Get()
							}
							result = result.UniqAppendProxyList(newResult)
						}
					}
				}
				break
			}
		}
	}
	return result
}
func GetUrls(url string) []string {
	resp, err := tool.GetHttpClient().Get(url)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}
	}
	text := string(body)
	return urlRe.FindAllString(text, -1)
}
func (w *WebFuzzClashNode) Get2ChanWG(pc chan proxy.Proxy, wg *sync.WaitGroup) {
	defer wg.Done()
	nodes := w.Get()
	log.Infoln("STATISTIC: WebFuzzClashNode \tcount=%d\turl=%s", len(nodes), w.Url)
	for _, node := range nodes {
		pc <- node
	}
}

func WebFuzzClashNodeGetter(options tool.Options) (getter Getter, err error) {
	urlInterface, found := options["url"]
	if found {
		url, err := AssertTypeStringNotNull(urlInterface)
		if err != nil {
			return nil, err
		}
		return &WebFuzzClashNode{Url: url}, nil
	}
	return nil, ErrorUrlNotFound
}
