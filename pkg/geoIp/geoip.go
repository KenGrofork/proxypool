package geoIp

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/vrichv/proxypool/config"
	"github.com/oschwald/geoip2-golang"
)

var GeoIpDB GeoIP

func InitGeoIpDB() error {
	parentPath := config.ResourceRoot()
	geodbPath := "assets/GeoLite2-City.mmdb"
	asnDbPath := "assets/GeoLite2-ASN.mmdb" // Add path for ASN database
	flagsPath := "assets/flags.json"
	geodb := filepath.Join(parentPath, geodbPath)
	asnDb := filepath.Join(parentPath, asnDbPath) // Add ASN database path
	flags := filepath.Join(parentPath, flagsPath)

	// 判断文件是否存在
	if _, err := os.Stat(geodb); err != nil && os.IsNotExist(err) {
		log.Println("文件不存在, 请自行下载 Geoip2 City 库, 并保存在", geodb)
		panic(err)
	}

	if _, err := os.Stat(asnDb); err != nil && os.IsNotExist(err) {
		log.Println("文件不存在, 请自行下载 Geoip2 ASN 库, 并保存在", asnDb)
		panic(err)
	}

	GeoIpDB = NewGeoIP(geodb, asnDb, flags) // Update NewGeoIP call
	return nil
}

// GeoIP2
type GeoIP struct {
	db       *geoip2.Reader
	asnDb    *geoip2.Reader // Add ASN database reader
	emojiMap map[string]string
}

type CountryEmoji struct {
	Code  string `json:"code"`
	Emoji string `json:"emoji"`
}

// new geoip from db file
func NewGeoIP(geodb, asnDb, flags string) (geoip GeoIP) {
	// 运行到这里时geodb只能为存在
	db, err := geoip2.Open(geodb)
	if err != nil {
		log.Fatal(err)
	}
	geoip.db = db

	// Open ASN database
	asnDB, err := geoip2.Open(asnDb)
	if err != nil {
		log.Fatal(err)
	}
	geoip.asnDb = asnDB

	_, err = os.Stat(flags)
	if err != nil && os.IsNotExist(err) {
		log.Println("flags 文件不存在, 请自行下载 flags.json, 并保存在 ", flags)
		os.Exit(1)
	} else {
		data, err := os.ReadFile(flags)
		if err != nil {
			log.Fatal(err)
			return
		}
		var countryEmojiList = make([]CountryEmoji, 0)
		err = json.Unmarshal(data, &countryEmojiList)
		if err != nil {
			log.Fatalln(err.Error())
			return
		}

		emojiMap := make(map[string]string)
		for _, i := range countryEmojiList {
			emojiMap[i.Code] = i.Emoji
		}
		geoip.emojiMap = emojiMap
	}
	return
}

// find ip info
func (g GeoIP) Find(ipORdomain string, isGetAsn bool) (ip, country, asnOrg string, err error) {
	ips, err := net.LookupIP(ipORdomain)
	if err != nil {
		return "", "🏁ZZ", "", err
	}
	ip = ips[0].String()
	var record *geoip2.City
	record, err = g.db.City(ips[0])
	if err != nil {
		return ip, "🏁ZZ", "", err
	}
	countryIsoCode := strings.ToUpper(record.Country.IsoCode)
	emoji, found := g.emojiMap[countryIsoCode]
	if found {
		country = fmt.Sprintf("%v%v", emoji, countryIsoCode)
	} else {
		country = "🏁ZZ"
	}

	if isGetAsn {
		// Get ASN information
		asnRecord, err := g.asnDb.ASN(ips[0])
		if err != nil {
			asnOrg = "" 
		} else {
			asnOrg = strings.ReplaceAll(asnRecord.AutonomousSystemOrganization, " ", "")
		}
	} else {
		asnOrg = ""
	}

	return ip, country, asnOrg, err
}