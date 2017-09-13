package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	hclog "github.com/brutella/hc/log"
	_ "github.com/joho/godotenv/autoload"
	"github.com/oleksandr/bonjour"
	"github.com/zubairhamed/canopus"
)

func browseForTradfriHub() (net.IP, int) {
	resolver, err := bonjour.NewResolver(nil)
	if err != nil {
		log.Panic(err)
	}

	results := make(chan *bonjour.ServiceEntry)
	err = resolver.Browse("_coap._udp", "local.", results)
	if err != nil {
		log.Fatal(err)
	}
	res := <-results
	log.Info(res)
	resolver.Exit <- true
	return res.AddrIPv4, res.Port
}

func (hub *tradfriHub) attachHcHandlers(bulbID string, acc *tradfriBulb) {
	acc.Lightbulb.On.OnValueRemoteUpdate(func(on bool) {
		hub.setBulbPower(bulbID, on)
		log.Info("Light state: %s", on)
	})
	acc.Lightbulb.Brightness.OnValueRemoteUpdate(func(brightness int) {
		hub.setBulbBrightness(bulbID, brightness)
		log.Info("Light brightness: %d", brightness)
	})
	acc.Lightbulb.ColorTemperature.OnValueRemoteUpdate(func(temp int) {
		hub.setBulbTemperature(bulbID, temp)
		log.Info("Light temp: %d", temp)
	})
}

func main() {
	ip, port := browseForTradfriHub()
	hub := initTradfriHub(ip, port)

	hclog.Debug.Enable()
	info := accessory.Info{
		Name: "Testlamp",
	}
	acc := newTradfriBulb(info)
	info2 := accessory.Info{
		Name: "Testlamp2",
	}
	acc2 := newTradfriBulb(info2)
	hub.attachHcHandlers("65538", acc)
	hub.attachHcHandlers("65537", acc2)

	config := hc.Config{Pin: "12344321", StoragePath: "./db"}
	t, err := hc.NewIPTransport(config, acc.Accessory, acc2.Accessory)
	if err != nil {
		log.Panic(err)
	}

	go t.Start()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Interrupt)

	for {
		select {

		case <-sigCh:
			t.Stop()
			os.Exit(0)
		}
	}

}

type tradfriHub struct {
	host string
	conn canopus.Connection
}

func initTradfriHub(ip net.IP, port int) *tradfriHub {
	hub := tradfriHub{}
	hub.host = fmt.Sprintf("%s:%d", ip, port)
	log.Infof("Connecting to gateway on %s...", hub.host)
	var err error
	hub.conn, err = canopus.DialDTLS(hub.host, "Client_identity", os.Getenv("TRADFRI_PSK"))
	if err != nil {
		log.Panic(err)
	}

	return &hub
}

func (hub *tradfriHub) getBulbStatus(id string) tradfriCOAPResponse {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Get).(*canopus.CoapRequest)
	req.SetRequestURI("/15001/" + id)

	resp, err := hub.conn.Send(req)
	if err != nil {
		log.Panic(err)
	}
	var res tradfriCOAPResponse
	json.Unmarshal(resp.GetMessage().GetPayload().GetBytes(), &res)
	return res
}

func (hub *tradfriHub) setBulb(bulbID, data string) {
	req := canopus.NewRequest(canopus.MessageConfirmable, canopus.Put).(*canopus.CoapRequest)
	req.SetRequestURI("/15001/" + bulbID)
	req.SetStringPayload(data)
	_, err := hub.conn.Send(req)
	if err != nil {
		log.Error(err)
	}
}

func (hub *tradfriHub) setBulbPower(bulbID string, on bool) {
	var onInt int
	if on {
		onInt = 1
	}
	hub.setBulb(bulbID, fmt.Sprintf(`{
		"3311": [{
			"5850": %d
		}]
	}`, onInt))
}

func (hub *tradfriHub) setBulbBrightness(id string, b int) {
	hub.setBulb(id, fmt.Sprintf(`{
		"3311": [{
			"5851": %d
		}]
	}`, b))
}

func (hub *tradfriHub) setBulbTemperature(id string, t int) {
	var color string
	if t < 200 { // whitest
		color = "f5faf6"
	} else if t < 300 {
		color = "f1e0b5"
	} else { // reddest
		color = "efd275"
	}

	log.Info("setting to: " + color)

	hub.setBulb(id, fmt.Sprintf(`{
		"3311": [{
			"5706": "%s"
		}]
	}`, color))
}
