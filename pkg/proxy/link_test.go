package proxy

import (
	"fmt"
	"github.com/vrichv/proxypool/pkg/tool"
	"github.com/ghodss/yaml"
	"testing"
)

func TestSSLink(t *testing.T) {
	ss, err := ParseSSLink("ss://YWVzLTI1Ni1jZmI6ZUlXMERuazY5NDU0ZTZuU3d1c3B2OURtUzIwMXRRMERAMTcyLjEwNC4xNjEuNTQ6ODA5OQ==#翻墙党223.13新加坡")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ss)
	fmt.Println(ss.Link())
	ss, err = ParseSSLink(ss.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ss)
}

func TestSSRLink(t *testing.T) {
	ssr, err := ParseSSRLink("ssr://MTcyLjEwNC4xNjEuNTQ6ODA5OTpvcmlnaW46YWVzLTI1Ni1jZmI6cGxhaW46WlVsWE1FUnVhelk1TkRVMFpUWnVVM2QxYzNCMk9VUnRVekl3TVhSUk1FUT0vP29iZnNwYXJhbT0mcHJvdG9wYXJhbT0mcmVtYXJrcz01Ny03NWFLWjVZV2FNakl6TGpFejVwYXc1WXFnNVoyaCZncm91cD01cGF3NVlxZzVaMmg=")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ssr)
	fmt.Println(ssr.Link())
	ssr, err = ParseSSRLink(ssr.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ssr)
	fmt.Println(ssr.ToClash())
}

func TestTrojanLink(t *testing.T) {
	trojan, err := ParseTrojanLink("trojan://AAA@example.com:33714?type=grpc&security=tls&serviceName=%2Fkjzhcuifg%2F&sni=example.com")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(trojan)
	fmt.Println(trojan.Link())
	trojan, err = ParseTrojanLink(trojan.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(trojan)
}

func TestVmessLink(t *testing.T) {
	//v, err := ParseVmessLink("vmess://ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuW+ruS/oeWFrOS8l+WPtyDlpJrlvannmoTlpKfljYPkuJbnlYwiLA0KICAiYWRkIjogInMyNzEuc25vZGUueHl6IiwNCiAgInBvcnQiOiAiNDQzIiwNCiAgImlkIjogIjZhOTAwZDYzLWNiOTItMzVhMC1hZWYwLTNhMGMxMWFhODUyMyIsDQogICJhaWQiOiAiMSIsDQogICJuZXQiOiAid3MiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAiczI3MS5zbm9kZS54eXoiLA0KICAicGF0aCI6ICIvcGFuZWwiLA0KICAidGxzIjogInRscyINCn0=")
	//v, err := ParseVmessLink("vmess://YXV0bzphMjA1ZjRiNi0xMzg2LTQ3NjUtYjQ0YS02YjFiYmE0N2Q1MzdAMTQyLjQuMTA0LjIyNjo0NDM?remarks=%F0%9F%87%BA%F0%9F%87%B8%20US_616%20caicai&obfsParam=www.036452916.xyz&path=/footers&obfs=websocket&tls=1&allowInsecure=1&alterId=64")
	v, err := ParseVmessLink("vmess://YXV0bzo1YjQ1ZjQ2Yi1iNTVmLTRkNWQtOGJjOS1jZjY1MzZlZjkyMzhAMTM3LjE3NS4zNS4xMzo0NDM?remarks=%F0%9F%87%BA%F0%9F%87%B8%20US_480%20caicai&obfsParam=www.4336705.xyz&path=/footers&obfs=websocket&tls=1&allowInsecure=1&alterId=64")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(v)
	fmt.Println(v.Link())
	v, err = ParseVmessLink(v.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(v)
}

func TestNewVmessParser(t *testing.T) {
	linkPayload := "ew0KICAidiI6ICIyIiwNCiAgInBzIjogIuW+ruS/oeWFrOS8l+WPtyDlpJrlvannmoTlpKfljYPkuJbnlYwiLA0KICAiYWRkIjogInMyNzEuc25vZGUueHl6IiwNCiAgInBvcnQiOiAiNDQzIiwNCiAgImlkIjogIjZhOTAwZDYzLWNiOTItMzVhMC1hZWYwLTNhMGMxMWFhODUyMyIsDQogICJhaWQiOiAiMSIsDQogICJuZXQiOiAid3MiLA0KICAidHlwZSI6ICJub25lIiwNCiAgImhvc3QiOiAiczI3MS5zbm9kZS54eXoiLA0KICAicGF0aCI6ICIvcGFuZWwiLA0KICAidGxzIjogInRscyINCn0="
	payload, err := tool.Base64DecodeString(linkPayload)
	if err != nil {
		fmt.Println("vmess link payload parse failed")
		return
	}
	jsonMap, err := str2jsonDynaUnmarshal(payload)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	vmessJson, err := mapStrInter2VmessLinkJson(jsonMap)
	if err != nil {
		panic(err)
	}
	fmt.Println(vmessJson)
}

func TestSSRClashYaml(t *testing.T) {
	str := "{\"name\":\"JP_609\",\"server\":\"13.231.143.248\",\"ip\":\"13.231.143.248\",\"outip\":\"\",\"port\":857,\"type\":\"ssr\",\"country\":\"JP\",\"flag\":\"🇯🇵\",\"usable\":true,\"delay\":847,\"Download\":0,\"Upload\":0,\"password\":\"CF5IKQ\",\"cipher\":\"chacha20-ietf\",\"protocol\":\"auth_aes128_sha1\",\"protocol-param\":\"45063:tyaGuO\",\"obfs\":\"tls1.2_ticket_auth\",\"obfs-param\":\"ffb1945063.microsoft.com\",\"group\":\"proxycrawler-clash\"}"
	var ssr ShadowsocksR
	err := yaml.Unmarshal([]byte(str), &ssr)
	if err != nil {
		panic(err)
	}
	fmt.Println(ssr)
	fmt.Println(ssr.Link())
	fmt.Println(ssr.ToClash())
	ssrp, err := ParseSSRLink(ssr.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(ssrp)
	fmt.Println(ssrp.ToClash())
}

func TestHysteria2Link(t *testing.T) {
	hysteria2, err := ParseHysteria2Link("hysteria2://letmein@example.com/?insecure=1&obfs=salamander&obfs-password=gawrgura&pinSHA256=deadbeef&sni=real.example.com")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hysteria2)
	fmt.Println(hysteria2.Link())
	hysteria2, err = ParseHysteria2Link(hysteria2.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hysteria2)
}

func TestVlessLink(t *testing.T) {
	vless, err := ParseVlessLink("vless://4843ddbd-8e3b-4199-a2cb-139dfxxc0aae56@example.com:30589?type=grpc&security=reality&serviceName=xcvsdfs&sni=www.speedtest.org&pbk=UtL7E0Gmxj3X5xJdcPAutpTRKo7K2hugkR0vwk2XroUM&fp=chrome")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(vless)
	fmt.Println(vless.Link())
	vless, err = ParseVlessLink(vless.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(vless)
}

func TestHysteriaLink(t *testing.T) {
	hysteria, err := ParseHysteriaLink("hysteria://host:1023?protocol=udp&auth=123456&peer=sni.domain&insecure=1&upmbps=100&downmbps=100&alpn=hysteria&obfs=xplus&obfsParam=123456#remarks")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hysteria)
	fmt.Println(hysteria.Link())
	hysteria, err = ParseHysteriaLink(hysteria.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(hysteria)
}

func TestTuicLink(t *testing.T) {
	trojan, err := ParseTuicLink("tuic://260139d2-8f99-4033-ab35-651925434420:ckiakmDojtl1@8.8.8.8:443?sni=jpnat.ddns.gay&alpn=h3&congestion_control=bbr#tuic_jp")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(trojan)
	fmt.Println(trojan.Link())
	trojan, err = ParseTuicLink(trojan.Link())
	if err != nil {
		t.Error(err)
	}
	fmt.Println(trojan)
}
