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

	o["ID"] = d.ID
	o["Name"] = d.Name

	if d.Lock != nil {
		o["Lock"] = "supported"
	} else {
		o["Lock"] = "unsupported"
	}

	if d.Mag != nil {
		s, err := ifk.Outputs[*d.Mag].State()
		if err != nil {
			o["Mag"] = "error"
		} else if s {
			o["Mag"] = "engaged"
		} else {
			o["Mag"] = "disengaged"
		}
	} else {
		o["Mag"] = "unsupported"
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
