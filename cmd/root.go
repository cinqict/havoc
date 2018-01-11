package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	sonos "github.com/ianr0bkny/go-sonos"
	"github.com/ianr0bkny/go-sonos/ssdp"
	"github.com/ianr0bkny/go-sonos/upnp"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "havoc",
}

var Shout = &cobra.Command{
	Use: "shout",
	Run: RunShout,
}

// vars for flags
var volume uint16
var wg sync.WaitGroup
var interf string
var baseBucket string

// var outputFile string

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().Uint16VarP(&volume, "vol", "v", 50, "shout volume, goes from 0 to 100")
	RootCmd.PersistentFlags().StringVarP(&interf, "interface", "i", "en0", "network interface to use for scan")
	RootCmd.PersistentFlags().StringVarP(&baseBucket, "basebucket", "b", "https://s3.eu-central-1.amazonaws.com/plpwavtest/", "basebucket")

	RootCmd.AddCommand(Shout)

}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	getSamplesFromS3()
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func RunShout(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		fmt.Printf("expecting one argument found %d", len(args))
	}

	smp := getSampleFileName(args[0])

	mgr := ssdp.MakeManager()

	// Discover()
	//  eth0 := Network device to query for UPnP devices
	// 11209 := Free local port for discovery replies
	// false := Do not subscribe for asynchronous updates
	mgr.Discover(interf, "11209", false)

	// SericeQueryTerms
	// A map of service keys to minimum required version
	qry := ssdp.ServiceQueryTerms{
		ssdp.ServiceKey("schemas-upnp-org-MusicServices"): -1,
	}
	// Look for the service keys in qry in the database of discovered devices

	result := mgr.QueryServices(qry)
	if devlist, has := result["schemas-upnp-org-MusicServices"]; has {
		wg.Add(len(devlist))

		for i, dev := range devlist {
			fmt.Println(i)

			go runSample(smp, volume, dev)

		}
	}

	// s := sonos.Connect(d, nil, sonos.SVC_AV_TRANSPORT)
	// for y, _ := range [10]struct{}{} {
	// 	fmt.Println(y)
	// 	spew.Dump(s.GetMediaInfo(0))
	// }
	wg.Wait()

	mgr.Close()
}

func runSample(f string, v uint16, d ssdp.Device) {

	fmt.Println("staring the sample run on" + d.Name())

	defer wg.Done()
	s := sonos.Connect(d, nil, sonos.SVC_AV_TRANSPORT)
	xs := sonos.Connect(d, nil, sonos.SVC_RENDERING_CONTROL)

	fmt.Println("connected to " + d.Name())

	cv, _ := xs.GetVolume(0, "Master")
	cmi, _ := s.GetMediaInfo(0)

	fmt.Println("saving current queue")

	fmt.Println("dropping sample on " + d.Name())

	e := s.SetAVTransportURI(0, f, "")
	if e != nil {
		fmt.Println(e)
	}

	e = xs.SetVolume(0, "Master", v)
	if e != nil {
		fmt.Println(e)
	}

	s.Play(0, upnp.PlaySpeed_1)

	mi, _ := s.GetPositionInfo(0)

	n, _ := getSecond(mi.TrackDuration)

	t := time.Now()

	for range [1000]struct{}{} {
		time.Sleep(100 * time.Millisecond)
		if int(time.Since(t).Seconds()) > n {
			break
		}

	}

	fmt.Println("done playing sample on " + d.Name())

	e = s.SetAVTransportURI(0, cmi.CurrentURI, cmi.CurrentURIMetaData)
	if e != nil {
		fmt.Println(e)
	}

	e = xs.SetVolume(0, "Master", cv)

	if e != nil {
		fmt.Println(e)
	}

	s.Play(0, upnp.PlaySpeed_1)

}

func getSecond(t string) (int, error) {
	return strconv.Atoi(strings.Split(t, ":")[2])
}
