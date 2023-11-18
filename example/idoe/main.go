package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

func main() {
	// _, _ = speedtest.FetchUserInfo()
	// Get a list of servers near a specified location
	// user.SetLocationByCity("Tokyo")
	// user.SetLocation("Osaka", 34.6952, 135.5006)

	// Select a network card as the data interface.
	// speedtest.WithUserConfig(&speedtest.UserConfig{Source: "192.168.1.101"})(speedtestClient)

	// Search server using serverID.
	// eg: fetch server with ID 28910.
	// speedtest.ErrEmptyServers will be returned if the server cannot be found.
	// server, err := speedtest.FetchServerByID("28910")

	//serverList, _ := speedtest.FetchServers()
	//targets, _ := serverList.FindServer([]int{})

	s, _ := speedtest.CustomServer("https://103.235.247.155:31555/upload.php")
	s.Context.IdoeAuthType = "ext_authz"
	s.Context.IdoeToken = "lxleeIdoetokenxxxxx"
	if s != nil {
		fmt.Println("type:", s.Context.IdoeAuthType, "token:", s.Context.IdoeToken)
		fmt.Printf("user: %v", s.Context.User)
	}
	//timeout := time.Duration(5 * time.Second)
	//s.TestDuration.Download = &timeout
	//svr.Context = &speedtest.Speedtest{

	//}

	//for _, s := range targets {

	// Please make sure your host can access this test server,
	// otherwise you will get an error.
	// It is recommended to replace a server at this time
	//checkError(s.PingTest(nil))
	//checkError(s.DownloadTest())
	//checkError(s.UploadTest())
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	checkError(s.DownloadTestContext(ctx))
	checkError(s.UploadTestContext(ctx))

	fmt.Printf("Latency: %s, Download: %f, Upload: %f\n", s.Latency, s.DLSpeed, s.ULSpeed)
	cancel()
	s.Context.Reset()
	//}
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}