package proxy

import (
	"encoding/json"
	"errors"
	"github.com/vrichv/proxypool/pkg/alpn"
	"github.com/vrichv/proxypool/pkg/geoIp"
	"strings"
	"unicode/utf8"
)

/* Base implements interface Proxy. It's the basic proxy struct. Vmess etc extends Base*/
type Base struct {
	Name    string `yaml:"name" json:"name" gorm:"index"`
	Server  string `yaml:"server" json:"server" gorm:"index"`
	Type    string `yaml:"type" json:"type" gorm:"index"`
	Country string `yaml:"country,omitempty" json:"country,omitempty" gorm:"index"`
	Port    int    `yaml:"port" json:"port" gorm:"index"`
	UDP     bool   `yaml:"udp,omitempty" json:"udp,omitempty"`
	Useable bool   `yaml:"useable,omitempty" json:"useable,omitempty" gorm:"index"`
}

// TypeName() Get specific proxy type
func (b *Base) TypeName() string {
	if b.Type == "" {
		return "unknown"
	}
	return b.Type
}

// SetName() to a proxy
func (b *Base) SetName(name string) {
	b.Name = name
}

func (b *Base) AddToName(name string) {
	b.Name = b.Name + name
}

func (b *Base) AddBeforeName(name string) {
	b.Name = name + b.Name
}

// SetIP() to a proxy
func (b *Base) SetIP(ip string) {
	b.Server = ip
}

// BaseInfo() get basic info struct of a proxy
func (b *Base) BaseInfo() *Base {
	return b
}

// Clone() returns a new basic proxy
func (b *Base) Clone() Base {
	c := *b
	return c
}

// SetUseable() set Base info "Useable" (true or false)
func (b *Base) SetUseable(useable bool) {
	b.Useable = useable
}

// SetUseable() set Base info "Country" (string)
func (b *Base) SetCountry(country string) {
	b.Country = country
}

type Proxy interface {
	String() string
	ToClash() string
	ToSurge() string
	Link() string
	Identifier() string
	SetName(name string)
	AddToName(name string)
	SetIP(ip string)
	TypeName() string //ss ssr vmess trojan
	BaseInfo() *Base
	Clone() Proxy
	SetUseable(useable bool)
	SetCountry(country string)
}

func ParseProxyFromLink(link string) (p Proxy, err error) {
	if strings.HasPrefix(link, "ssr://") {
		p, err = ParseSSRLink(link)
	} else if strings.HasPrefix(link, "vmess://") {
		p, err = ParseVmessLink(link)
	} else if strings.HasPrefix(link, "ss://") {
		p, err = ParseSSLink(link)
	} else if strings.HasPrefix(link, "trojan://") {
		p, err = ParseTrojanLink(link)
	} else if strings.HasPrefix(link, "trojan-go://") {
		p, err = ParseTrojanLink(link)
	} else if strings.HasPrefix(link, "hysteria2://") {
		p, err = ParseHysteria2Link(link)
	} else if strings.HasPrefix(link, "hy2://") {
		p, err = ParseHysteria2Link(link)
	} else if strings.HasPrefix(link, "vless://") {
		p, err = ParseVlessLink(link)
	} else if strings.HasPrefix(link, "vless1://") {
		p, err = ParseVlessLink(link)
	} else if strings.HasPrefix(link, "hysteria://") {
		p, err = ParseHysteriaLink(link)
	} else if strings.HasPrefix(link, "hy://") {
		p, err = ParseHysteriaLink(link)
	} else if strings.HasPrefix(link, "tuic://") {
		p, err = ParseTuicLink(link)
	}

	if err != nil || p == nil {
		return nil, errors.New("link parse failed")
	}
	_, country, _, err := geoIp.GeoIpDB.Find(p.BaseInfo().Server, false) // IP库不准
	if err != nil {
		country = "🏁ZZ"
	}
	p.SetCountry(country)
	// trojan依赖域名？<-这是啥?不管什么情况感觉都不应该替换域名为IP（主要是IP库的质量和节点质量不该挂钩）
	//if p.TypeName() != "trojan" {
	//	p.SetIP(ip)
	//}
	return
}

//	func isNumber(value interface{}) bool {
//		switch value.(type) {
//		case int, int8, int16, int32, int64:
//			return true
//		case uint, uint8, uint16, uint32, uint64:
//			return true
//		case float32, float64:
//			return true
//		default:
//			return false
//		}
//	}

func ParseProxyFromClashProxy(p map[string]interface{}) (proxy Proxy, err error) {
	p["name"] = ""
	//if p["password"] != nil {
	//	password, ok := p["password"].(string)
	//	if ok {
	//		if _, err := strconv.ParseFloat(password, 64); err == nil {
	//			return nil, errors.New("password is number")
	//		}
	//	} else {
	//		if isNumber(p["password"]) {
	//			return nil, errors.New("password is number")
	//		}
	//	}
	//}
	for key, value := range p {
		str, ok := value.(string)
		if !ok {
			continue
		}
		if strings.Contains(str, "%") {
			return nil, errors.New("clash json parse failed")
		}
		for _, runeValue := range str {
			// 检查字符是否为有效的UTF-8编码
			if !utf8.ValidRune(runeValue) || runeValue == utf8.RuneError {
				return nil, errors.New("clash json parse failed")
			}
		}
		for _, runeValue := range key {
			// 检查字符是否为有效的UTF-8编码
			if !utf8.ValidRune(runeValue) || runeValue == utf8.RuneError {
				return nil, errors.New("clash json parse failed")
			}
		}
	}

	pjson, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}
	if p["type"] == nil {
		return nil, errors.New("clash json parse failed")
	}
	switch p["type"].(string) {
	case "ss":
		var proxy Shadowsocks
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		//if !ValidPassword(&proxy.Password) {
		//	return nil, errors.New("Password Error")
		//}
		return &proxy, nil
	case "ssr":
		var proxy ShadowsocksR
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		//if !ValidPassword(&proxy.Password) {
		//	return nil, errors.New("Password Error")
		//}
		if !ValidParams(&proxy.ProtocolParam) {
			return nil, errors.New("Password Error")
		}
		if !ValidParams(&proxy.ObfsParam) {
			return nil, errors.New("Password Error")
		}
		return &proxy, nil
	case "vmess":
		var proxy Vmess
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		proxy.ALPN = alpn.FormatAlpnArray(proxy.ALPN)
		return &proxy, nil
	case "trojan":
		var proxy Trojan
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		//if !ValidPassword(&proxy.Password) {
		//	return nil, errors.New("Password Error")
		//}
		proxy.ALPN = alpn.FormatAlpnArray(proxy.ALPN)
		return &proxy, nil
	case "hysteria2":
		var proxy Hysteria2
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		//if !ValidPassword(&proxy.Password) {
		//	return nil, errors.New("Password Error")
		//}
		return &proxy, nil
	case "vless":
		var proxy Vless
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		if proxy.Flow == "xtls-rprx-direct" {
			return nil, errors.New("legacy XTLS protocol xtls-rprx-direct is deprecated and no longer supported")
		}
		//if !ValidPassword(&proxy.Password) {
		//	return nil, errors.New("Password Error")
		//}
		proxy.ALPN = alpn.FormatAlpnArray(proxy.ALPN)
		return &proxy, nil
	case "hysteria":
		var proxy Hysteria
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	case "tuic":
		var proxy Tuic
		err := json.Unmarshal(pjson, &proxy)
		if err != nil {
			return nil, err
		}
		return &proxy, nil
	}
	return nil, errors.New("clash json parse failed")
}

//func ValidPassword(pass interface{}) (flag bool) {
//	password, ok := pass.(string)
//	if ok {
//		if _, err := strconv.ParseFloat(password, 64); err == nil {
//			return false
//		}
//	} else {
//		var npassword string
//		if passStr, ok := pass.(*string); ok {
//			if passStr != nil {
//				npassword = *passStr
//				if _, err := strconv.ParseFloat(npassword, 64); err == nil {
//					return false
//				}
//			}
//		} else {
//			return false
//		}
//	}
//	return true
//}

func ValidParams(param interface{}) (flag bool) {
	str, ok := param.(string)
	if ok {
		if strings.Contains(str, "%") {
			return false
		}
		if strings.Contains(str, "\\") {
			return false
		}
		for _, runeValue := range str {
			// 检查字符是否为有效的UTF-8编码
			if !utf8.ValidRune(runeValue) || runeValue == utf8.RuneError {
				return false
			}
		}
		return true
	} else {
		return false
	}
}
func GoodNodeThatClashUnsupported(b Proxy) bool {
	switch b.TypeName() {
	case "ss":
		ss := b.(*Shadowsocks)
		if ss == nil {
			return false
		}
		if ss.Cipher == "none" {
			return true
		} else {
			return false
		}
	case "ssr":
		ssr := b.(*ShadowsocksR)
		if ssr == nil {
			return false
		}
		return true
	}
	return false
}
