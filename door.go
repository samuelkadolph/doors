package main

import (
	"encoding/json"
	"github.com/samuelkadolph/go/phidgets"
	"time"
)

type Door struct {
	ID   string
	Lock *int
	Mag  *int
	Name string
}

func (d *Door) MagDisengage(ifk *phidgets.InterfaceKit) error {
	return ifk.Outputs[*d.Mag].SetState(false)
}

func (d *Door) MagEngage(ifk *phidgets.InterfaceKit) error {
	return ifk.Outputs[*d.Mag].SetState(true)
}

func (d *Door) MarshalJSON() ([]byte, error) {
	o := make(map[string]interface{})

	o["id"] = d.ID
	o["name"] = d.Name

	if d.Lock != nil {
		o["lock"] = "supported"
	} else {
		o["lock"] = "unsupported"
	}

	if d.Mag != nil {
		s, err := ifk.Outputs[*d.Mag].State()
		if err != nil {
			o["mag"] = "error"
		} else if s {
			o["mag"] = "engaged"
		} else {
			o["mag"] = "disengaged"
		}
	} else {
		o["mag"] = "unsupported"
	}

	return json.Marshal(o)
}

func (d *Door) Open(ifk *phidgets.InterfaceKit) error {
	if err := ifk.Outputs[*d.Lock].SetState(true); err != nil {
		return err
	}

	time.Sleep(200 * time.Millisecond)

	return ifk.Outputs[*d.Lock].SetState(false)
}
