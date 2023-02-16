package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	showList     = kingpin.Flag("list", "Show available speedtest.net servers.").Short('l').Bool()
	serverIds    = kingpin.Flag("server", "Select server id to run speedtest.").Short('s').Ints()
	customURL    = kingpin.Flag("custom-url", "Specify the url of the server instead of getting a list from Speedtest.net").String()
	savingMode   = kingpin.Flag("saving-mode", "Using less memory (≒10MB), though low accuracy (especially > 30Mbps).").Bool()
	jsonOutput   = kingpin.Flag("json", "Output results in json format").Bool()
	location     = kingpin.Flag("location", "Change the location with a precise coordinate. Format: lat,lon").String()
	city         = kingpin.Flag("city", "Change the location with a predefined city label.").String()
	showCityList = kingpin.Flag("city-list", "List all predefined city labels.").Bool()
	proxy        = kingpin.Flag("proxy", "Set a proxy(http(s) or socks) for the speedtest.").String()
	source       = kingpin.Flag("source", "Set the source interface(tcp[4/6]://ip) for the speedtest.").String()
)

type fullOutput struct {
	Timestamp outputTime        `json:"timestamp"`
	UserInfo  *speedtest.User   `json:"user_info"`
	Servers   speedtest.Servers `json:"servers"`
}
type outputTime time.Time

func main() {
	kingpin.Version(speedtest.Version())
	kingpin.Parse()

	var speedtestClient = speedtest.New()

	if len(*proxy) > 0 || len(*source) > 0 {
		config := &speedtest.UserConfig{
			UserAgent: speedtest.DefaultUserAgent,
			Proxy:     *proxy,
			Source:    *source,
		}
		speedtest.WithUserConfig(config)(speedtestClient)
	}

	user, err := speedtestClient.FetchUserInfo()
	if err != nil {
		fmt.Printf("Warning: can not fetch user information. err: %v\n", err.Error())
		return
	}

	if *showCityList {
		speedtest.PrintCityList()
		return
	}

	if len(*city) > 0 {
		err = user.SetLocationByCity(*city)
		if err != nil {
			fmt.Printf("Warning: skipping command line arguments: --city. err: %v\n", err.Error())
		}
	}

	if len(*location) > 0 {
		err = user.ParseAndSetLocation(*location)
		if err != nil {
			fmt.Printf("Warning: skipping command line arguments: --location. err: %v\n", err.Error())
		}
	}

	if !*jsonOutput {
		showUser(user)
	}

	servers, err := speedtestClient.FetchServers(user)
	checkError(err)
	var targets speedtest.Servers
	if *customURL == "" {
		if *showList {
			showServerList(servers)
			return
		}

		targets, err = servers.FindServer(*serverIds)
		checkError(err)

	} else {
		target, err := speedtest.CustomServer(*customURL)
		checkError(err)
		targets = []*speedtest.Server{target}
	}

	startTest(targets, *savingMode, *jsonOutput)

	if *jsonOutput {
		jsonBytes, err := json.Marshal(
			fullOutput{
				Timestamp: outputTime(time.Now()),
				UserInfo:  user,
				Servers:   targets,
			},
		)
		checkError(err)

		fmt.Println(string(jsonBytes))
	}
}

func startTest(servers speedtest.Servers, savingMode bool, jsonOutput bool) {
	for _, s := range servers {
		if !jsonOutput {
			showServer(s)
		}

		err := s.PingTest()
		checkError(err)

		if jsonOutput {
			err := s.DownloadTest(savingMode)
			checkError(err)

			err = s.UploadTest(savingMode)
			checkError(err)

			continue
		}

		showLatencyResult(s)

		err = testDownload(s, savingMode)
		checkError(err)
		err = testUpload(s, savingMode)
		checkError(err)

		showServerResult(s)
	}

	if !jsonOutput && len(servers) > 1 {
		showAverageServerResult(servers)
	}
}

func testDownload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("Download Test: ")
	go dots(quit)
	err := server.DownloadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return err
}

func testUpload(server *speedtest.Server, savingMode bool) error {
	quit := make(chan bool)
	fmt.Printf("Upload Test: ")
	go dots(quit)
	err := server.UploadTest(savingMode)
	quit <- true
	if err != nil {
		return err
	}
	fmt.Println()
	return nil
}

func dots(quit chan bool) {
	for {
		select {
		case <-quit:
			return
		default:
			time.Sleep(time.Second)
			fmt.Print(".")
		}
	}
}

func showUser(user *speedtest.User) {
	if user.IP != "" {
		fmt.Printf("Testing From IP: %s\n", user.String())
	}
}

func showServerList(servers speedtest.Servers) {
	for _, s := range servers {
		fmt.Printf("[%4s] %8.2fkm ", s.ID, s.Distance)
		fmt.Printf(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
	}
}

func showServer(s *speedtest.Server) {
	fmt.Printf(" \n")
	fmt.Printf("Target Server: [%4s] %8.2fkm ", s.ID, s.Distance)
	fmt.Printf(s.Name + " (" + s.Country + ") by " + s.Sponsor + "\n")
}

func showLatencyResult(server *speedtest.Server) {
	fmt.Println("Latency:", server.Latency)
}

// ShowResult : show testing result
func showServerResult(server *speedtest.Server) {
	fmt.Printf(" \n")

	fmt.Printf("Download: %5.2f Mbit/s\n", server.DLSpeed)
	fmt.Printf("Upload: %5.2f Mbit/s\n\n", server.ULSpeed)
	valid := server.CheckResultValid()
	if !valid {
		fmt.Println("Warning: Result seems to be wrong. Please speedtest again.")
	}
}

func showAverageServerResult(servers speedtest.Servers) {
	avgDL := 0.0
	avgUL := 0.0
	for _, s := range servers {
		avgDL = avgDL + s.DLSpeed
		avgUL = avgUL + s.ULSpeed
	}
	fmt.Printf("Download Avg: %5.2f Mbit/s\n", avgDL/float64(len(servers)))
	fmt.Printf("Upload Avg: %5.2f Mbit/s\n", avgUL/float64(len(servers)))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (t outputTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", time.Time(t).Format("2006-01-02 15:04:05.000"))
	return []byte(stamp), nil
}
