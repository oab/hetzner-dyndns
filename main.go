package main

import (
	"errors"
	"fmt"
	"bytes"
	"io/ioutil"
	"time"
	"net"
	"net/http"
	"flag"
	"log"
	"github.com/pion/stun"
)


func getIP(server *string) net.IP {
    var ip net.IP = nil
	c, err := stun.Dial("udp4", *server)
	if err != nil {
		log.Fatal("dial:", err)
	}
	if err = c.Do(stun.MustBuild(stun.TransactionID, stun.BindingRequest), func(res stun.Event) {
		if res.Error != nil {
			log.Fatalln(res.Error)
		}
		var xorAddr stun.XORMappedAddress
		if getErr := xorAddr.GetFrom(res.Message); getErr != nil {
			log.Fatalln(getErr)
		}
		ip = xorAddr.IP
	}); err != nil {
		log.Fatal("do:", err)
	}
	if err := c.Close(); err != nil {
		log.Fatalln(err)
	}
	return ip
      	
}


func sendUpdateRecord(value, apitoken, ttl, recordtype, recordid, recordname, zoneid * string) {
	json := []byte(fmt.Sprintf(`{"value":%q,"ttl":%s,"type":%q,"name":%q,"zone_id":%q}`,*value,*ttl,*recordtype,*recordname,*zoneid))
	body := bytes.NewBuffer(json)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", fmt.Sprintf(`https://dns.hetzner.com/api/v1/records/%s`,*recordid), body)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Auth-API-Token", *apitoken)

	resp, err := client.Do(req)
	
	if err != nil {
		log.Println("Failure : ", err)
	}

	respBody, _ := ioutil.ReadAll(resp.Body)

	log.Println("response Status : ", resp.Status)
	log.Println("response Headers : ", resp.Header)
	log.Println("response Body : ", string(respBody))
}


var (
	stunserver  = flag.String("stunserver","","stunserver to use")
    refreshrate = flag.String("refreshrate","", "refresh rate for update")
    apitoken    = flag.String("apitoken","","hetzner rest api token")
    recordtype  = flag.String("recordtype","","record type")
    ttl         = flag.String("ttl","","time to live")
    recordname  = flag.String("recordname","","record name")
    recordid    = flag.String("recordid","","record id")
    zoneid      = flag.String("zoneid","","zone id")
)

func checkflag(flag *string, name string) {
	if *flag == "" {
	  err := fmt.Sprintf("%s must be specified",name)
	  log.Fatalln(errors.New(err))
	}	
}

func main() {
  flag.Parse()
  sleepduration,err := time.ParseDuration(*refreshrate)
  if err != nil {
   log.Fatalln(err); 		
  }
 
  checkflag(stunserver,"stunserver");
  checkflag(apitoken,"apitoken");
  checkflag(recordtype,"recordtype");
  checkflag(recordname,"recordname");
  checkflag(recordid,"recordid");
  checkflag(zoneid,"zoneid");
  checkflag(refreshrate,"refreshrate")
  checkflag(ttl,"ttl")
	
  cachedip := getIP(stunserver)
	
  for {
    log.Println("Cached external IP:", cachedip)
	if ip := getIP(stunserver); ip.Equal(cachedip) {			
      log.Println("Change detected, new IP is: ",ip)
	  ipstring := ip.String()
	  sendUpdateRecord(&ipstring,apitoken,ttl,recordtype,recordid,recordname,zoneid)
	  cachedip = ip
	}
			
	time.Sleep(sleepduration)
  }
  	
  
}
