package main

import (
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
)

type tradfriBulbService struct {
	*service.Service

	On               *characteristic.On
	Brightness       *characteristic.Brightness
	ColorTemperature *characteristic.ColorTemperature
}

func newTradfriBulbService() *tradfriBulbService {
	svc := tradfriBulbService{}
	svc.Service = service.New(service.TypeLightbulb)

	svc.On = characteristic.NewOn()
	svc.AddCharacteristic(svc.On.Characteristic)

	svc.Brightness = characteristic.NewBrightness()
	svc.AddCharacteristic(svc.Brightness.Characteristic)

	svc.ColorTemperature = characteristic.NewColorTemperature()
	svc.AddCharacteristic(svc.ColorTemperature.Characteristic)

	return &svc
}

type tradfriBulb struct {
	*accessory.Accessory
	Lightbulb *tradfriBulbService
}

func newTradfriBulb(info accessory.Info) *tradfriBulb {
	acc := tradfriBulb{}
	acc.Accessory = accessory.New(info, accessory.TypeLightbulb)
	acc.Lightbulb = newTradfriBulbService()

	acc.Lightbulb.Brightness.SetValue(100)
	acc.Lightbulb.On.SetValue(false)
	acc.Lightbulb.ColorTemperature.SetValue(200)

	acc.AddService(acc.Lightbulb.Service)

	return &acc
}

type tradfriCOAPResponse struct {
	Name    string `json:"9001"`
	Control struct {
		On         int    `json:"5850"`
		Brightness int    `json:"5851"`
		Color      string `json:"5706"`
	} `json:"3311"`
}
