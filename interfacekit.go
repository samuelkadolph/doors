package main

import (
	"github.com/samuelkadolph/go/phidgets"
	"github.com/samuelkadolph/go/phidgets/raw"
	"time"
)

type InterfaceKit struct {
	AttachmentTimeout *int
	Doors             []*Door
	Host              *string
	Label             *string
	LockDelay         *int
	Password          *string
	Port              *int
	Serial            *int

	*phidgets.InterfaceKit
}

func (i *InterfaceKit) Load() error {
	var err error

	if i.InterfaceKit, err = phidgets.NewInterfaceKit(); err != nil {
		return err
	}

	if err = i.Open(i.connector()); err != nil {
		return err
	}

	if err = i.WaitForAttachment(i.timeout()); err != nil {
		return err
	}

	for _, d := range i.Doors {
		d.ifk = i
	}

	return nil
}

func (i *InterfaceKit) connector() phidgets.Connector {
	var c phidgets.Connector
	var host string
	var password string
	var port int

	if i.Password != nil {
		password = *i.Password
	}

	if i.Port != nil {
		port = *i.Port
	} else {
		port = 5001
	}

	if i.Host != nil {
		host = *i.Host
		if i.Serial != nil {
			c = phidgets.RemoteIPSerial{*i.Serial, host, port, password}
		} else if i.Label != nil {
			c = phidgets.RemoteIPLabel{*i.Label, host, port, password}
		} else {
			c = phidgets.RemoteIPSerial{raw.Any, host, port, password}
		}
	} else {
		if i.Serial != nil {
			c = phidgets.Serial{*i.Serial}
		} else if i.Label != nil {
			c = phidgets.Label{*i.Label}
		} else {
			c = phidgets.Any
		}
	}

	return c
}

func (i *InterfaceKit) timeout() time.Duration {
	if i.AttachmentTimeout != nil {
		return time.Duration(*i.AttachmentTimeout) * time.Millisecond
	}

	return 2000 * time.Millisecond
}
