package proxy

import (
	"encoding/json"
	"errors"
	A "github.com/vrichv/proxypool/pkg/alpn"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrorNotVlessLink = errors.New("not a correct vless link")
)

type RealityOpts struct {
	PublicKey string `yaml:"public-key,omitempty" json:"public-key,omitempty"`
	ShortId   string `yaml:"short-id,omitempty" json:"short-id,omitempty"`
}

type Vless struct {
	Base
	UUID              string       `yaml:"uuid" json:"uuid"`
	Network           string       `yaml:"network,omitempty" json:"network,omitempty"`
	TLS               bool         `yaml:"tls,omitempty" json:"tls,omitempty"`
	SkipCertVerify    bool         `yaml:"skip-cert-verify,omitempty" json:"skip-cert-verify,omitempty"`
	Flow              string       `yaml:"flow,omitempty" json:"flow,omitempty"`
	FingerPrint       string       `yaml:"fingerprint,omitempty" json:"fingerprint,omitempty"`
	XUDP              bool         `yaml:"xudp,omitempty" json:"xudp,omitempty"`
	ClientFingerprint string       `yaml:"client-fingerprint,omitempty" json:"client-fingerprint,omitempty"`
	ServerName        string       `yaml:"servername,omitempty" json:"servername,omitempty"`
	RealityOpts       *RealityOpts `yaml:"reality-opts,omitempty" json:"reality-opts,omitempty"`
	GrpcOpts          *GrpcOptions `yaml:"grpc-opts,omitempty" json:"grpc-opts,omitempty"`
	WSOpts            *WSOptions   `yaml:"ws-opts,omitempty" json:"ws-opts,omitempty"`
	ALPN              []string     `yaml:"alpn,omitempty" json:"alpn,omitempty"`
	PacketEncoding    string       `yaml:"packet-encoding" json:"packet-encoding"`
}

func (v Vless) String() string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func (v Vless) ToClash() string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return "- " + string(data)
}

func (v Vless) ToSurge() string {
	return ""
}

func (v Vless) Link() string {
	query := url.Values{}
	if v.Flow != "" {
		query.Set("flow", url.QueryEscape(v.Flow))
	}
	if v.TLS {
		query.Set("tls", url.QueryEscape("true"))
	}
	if v.WSOpts != nil && v.WSOpts.Path != "" {
		query.Set("path", url.QueryEscape(v.WSOpts.Path))
	}

	if v.GrpcOpts != nil && v.GrpcOpts.GrpcServiceName != "" {
		query.Set("serviceName", url.QueryEscape(v.GrpcOpts.GrpcServiceName))
	}

	if v.SkipCertVerify {
		query.Set("allowInsecure", url.QueryEscape("true"))
	}
	if v.FingerPrint != "" {
		query.Set("fp", url.QueryEscape(v.FingerPrint))
	}
	if v.ServerName != "" {
		query.Set("sni", url.QueryEscape(v.ServerName))
	}
	if v.RealityOpts != nil {
		query.Set("security", url.QueryEscape("reality"))
		if v.RealityOpts.PublicKey != "" {
			query.Set("pbk", url.QueryEscape(v.RealityOpts.PublicKey))
		}
		if v.RealityOpts.ShortId != "" {
			query.Set("sid", url.QueryEscape(v.RealityOpts.ShortId))
		}
	}
	if v.Network != "" {
		query.Set("type", url.QueryEscape(v.Network))
	}
	if len(v.ALPN) != 0 {
		query.Set("alpn", url.QueryEscape(v.ALPN[0]))
	}
	uri := url.URL{
		Scheme:   "vless",
		User:     url.User(url.QueryEscape(v.UUID)),
		Host:     net.JoinHostPort(v.Server, strconv.Itoa(v.Port)),
		RawQuery: query.Encode(),
		Fragment: v.Name,
	}

	return uri.String()
}

func (v Vless) Identifier() string {
	data, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(data)
}

func (v Vless) Clone() Proxy {
	return &v
}

func ParseVlessLink(link string) (*Vless, error) {
	if !strings.HasPrefix(link, "vless://") && !strings.HasPrefix(link, "vless1://") {
		return nil, ErrorNotVlessLink
	}
	uri, err := url.Parse(link)
	if err != nil {
		return nil, ErrorNotSSLink
	}

	password := uri.User.Username()
	password, _ = url.QueryUnescape(password)

	server := uri.Hostname()
	port, _ := strconv.Atoi(uri.Port())

	moreInfos := uri.Query()
	sni := moreInfos.Get("sni")
	sni, snierr := url.QueryUnescape(sni)
	transformType := moreInfos.Get("type")
	transformType, _ = url.QueryUnescape(transformType)
	host := moreInfos.Get("host")
	host, hostErr := url.QueryUnescape(host)
	path := moreInfos.Get("path")
	path, patherr := url.QueryUnescape(path)
	serviceName := moreInfos.Get("serviceName")
	serviceName, serviceNameerr := url.QueryUnescape(serviceName)
	flow := moreInfos.Get("flow")
	flow, flowerr := url.QueryUnescape(flow)
	security := moreInfos.Get("security")
	security, securityerr := url.QueryUnescape(security)
	pbk := moreInfos.Get("pbk")
	pbk, _ = url.QueryUnescape(pbk)
	sid := moreInfos.Get("sid")
	sid, siderr := url.QueryUnescape(sid)
	allowInsecure := moreInfos.Get("allowInsecure")
	allowInsecure, allowInsecureerr := url.QueryUnescape(allowInsecure)
	fingerprint := moreInfos.Get("fp")
	fingerprint, fingerprinterr := url.QueryUnescape(fingerprint)
	udp := moreInfos.Get("udp")
	udp, udperr := url.QueryUnescape(udp)
	alpn := moreInfos.Get("alpn")
	alpn, _ = url.QueryUnescape(alpn)

	if port == 0 {
		port = 443
	}
	t := &Vless{
		Base: Base{
			Name:   "",
			Server: server,
			Port:   port,
			Type:   "vless",
		},
		Network: "tcp",
	}
	if udperr == nil && udp == "false" {
		t.Base.UDP = false
	} else {
		t.Base.UDP = true
	}
	if transformType == "ws" {
		t.Network = "ws"
	} else if transformType == "grpc" {
		t.Network = "grpc"
	} else if transformType == "tcp" {
		t.Network = "tcp"
	} else {
		return nil, ErrorNotVlessLink
	}
	if securityerr == nil {
		if security == "tls" {
			t.SkipCertVerify = false
		} else if security == "reality" {
			if siderr == nil && sid != "" {
				t.RealityOpts = &RealityOpts{
					PublicKey: pbk,
					ShortId:   sid,
				}
			} else {
				t.RealityOpts = &RealityOpts{
					PublicKey: pbk,
				}
			}
		} else {
			t.SkipCertVerify = true
		}
	}
	if snierr == nil && sni != "" {
		t.ServerName = sni
	}
	if fingerprinterr == nil && fingerprint != "" {
		t.FingerPrint = fingerprint
	}
	if pbk != "" || flow == "xtls-rprx-vision" {
		t.ClientFingerprint = "chrome"
	}
	if allowInsecureerr == nil && allowInsecure == "true" {
		t.SkipCertVerify = true
	}
	if flowerr == nil && flow != "" {
		if flow == "xtls-rprx-direct" {
			return nil, ErrorNotSSLink
		}
		t.Flow = flow
	}
	if serviceNameerr == nil && serviceName != "" {
		t.GrpcOpts = &GrpcOptions{
			GrpcServiceName: serviceName,
		}
	}
	if alpn != "" {
		t.ALPN = []string{alpn}
		t.ALPN = A.FormatAlpnArray(t.ALPN)
	}
	if patherr == nil && path != "" {
		if hostErr == nil && host != "" {
			t.ServerName = host
			wsHeaders := make(map[string]string)
			wsHeaders["Host"] = host
			t.WSOpts = &WSOptions{
				Path:    path,
				Headers: wsHeaders,
			}
		} else {
			t.ServerName = server
			wsHeaders := make(map[string]string)
			wsHeaders["Host"] = server
			t.WSOpts = &WSOptions{
				Path:    path,
				Headers: wsHeaders,
			}
		}
	}
	return t, nil
}

var (
	vlessPlainRe = regexp.MustCompile("vless://([A-Za-z0-9+/_&?=@:%.-])+")
)

func GrepVlessLinkFromString(text string) []string {
	results := make([]string, 0)
	if !strings.Contains(text, "vless://") {
		return results
	}
	texts := strings.Split(text, "vless://")
	for _, text := range texts {
		results = append(results, vlessPlainRe.FindAllString("vless://"+text, -1)...)
	}
	return results
}
