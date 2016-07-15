package main

import (
	"flag"
	"fmt"
	"github.com/Unknwon/goconfig"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	flaga        bool = true
	configFile   string
	interval     int
	failureTimes int
	domain       string
	ips          []string
	ports        []string
	hosts        string = `127.0.0.1	localhost
127.0.1.1	vhdubuntu1404.cs1cloud.internal	vhdubuntu1404

# The following lines are desirable for IPv6 capable hosts
::1     localhost ip6-localhost ip6-loopback
ff02::1 ip6-allnodes
ff02::2 ip6-allrouters`
	loger *log.Logger
)

func init() {

	flag.StringVar(&configFile, "c", "check.conf", "configure file path")
	goConfigFile, err := goconfig.LoadConfigFile(configFile)
	printERR(err)
	domain, err = goConfigFile.GetValue("checkport", "domain")
	printERR(err)
	logpath, err := goConfigFile.GetValue("checkport", "logpath")
	printERR(err)
	logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_SYNC|os.O_RDWR, 0644)
	printERR(err)
	loger = log.New(logfile, "", log.LstdFlags)
	interval = goConfigFile.MustInt("checkport", "interval", 5)
	failureTimes = goConfigFile.MustInt("checkport", "failureTimes", 2)
	tmpstring, err := goConfigFile.GetValue("services", "ip")
	printERR(err)
	ips = strings.Split(tmpstring, ",")
	tmpport, err := goConfigFile.GetValue("services", "port")
	printERR(err)
	ports = strings.Split(tmpport, ",")
}
func main() {
	// 后台运行

	if os.Getppid() != 1 {
		command := exec.Command(os.Args[0], os.Args[1:]...)
		command.Stdout = os.Stdout
		command.Start()
		fmt.Println("[PID]", command.Process.Pid)
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	for {
		<-ticker.C
		for _, ip := range ips {
			record := getRecord()
			if record[0].String() == ip {
				for _, port := range ports {
					err := checkPort(ip + ":" + port)
					if err != nil {
						flaga = false
						loger.Println(ip + ":" + port + "  is down! " + "[error]: " + err.Error())

						break
					}
				}
			} else if !flaga {
				changeRecord(ip)
				flaga = true
			}
		}
	}
}

func checkPort(addr string) error {
	tmpconn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	defer tmpconn.Close()
	return err
}

func checkErr(err error) {
	if err != nil {
		loger.Println(err)
	}
}
func getRecord() (addrs []net.IP) {
	addrs, err := net.LookupIP(domain)
	checkErr(err)
	return addrs
}
func printERR(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
func changeRecord(ip string) {
	file, err := os.OpenFile("/etc/hosts", os.O_RDWR|os.O_CREATE, 0644)
	defer file.Close()
	checkErr(err)
	record := ip + "    " + domain
	fmt.Fprintf(file, "%s\n%s\n", hosts, record)
	loger.Println("change to ", ip)
}
