package main

import (
	"encoding/json"
	"sync"
	"time"
)

type Door struct {
	Floor        *string
	ID           string
	Lock         *int
	LockFeedback *int
	Mag          *int
	MagFeedback  *int
	Name         *string

	ifk       *InterfaceKit
	lockCond  *sync.Cond
	lockMutex *sync.Mutex
	magCond   *sync.Cond
	magMutex  *sync.Mutex
}

type HTTPDoor struct {
	Door
}

func (d *Door) LockStatus() string {
	var err error
	var s bool

	if d.LockFeedback != nil {
		s, err = d.ifk.Inputs[*d.LockFeedback].State()
	} else if d.Lock != nil {
		s, err = d.ifk.Outputs[*d.Lock].State()
	} else {
		return "unsupported"
	}

	if err != nil {
		return "error"
	} else if s {
		return "unlocked"
	}

	return "locked"
}

func (d *Door) MagDisengage() error {
	return d.ifk.Outputs[*d.Mag].SetState(false)
}

func (d *Door) MagEngage() error {
	return d.ifk.Outputs[*d.Mag].SetState(true)
}

func (d *Door) MagStatus() string {
	var err error
	var s bool

	if d.MagFeedback != nil {
		s, err = d.ifk.Inputs[*d.MagFeedback].State()
	} else if d.Mag != nil {
		s, err = d.ifk.Outputs[*d.Mag].State()
	} else {
		return "unsupported"
	}

	if err != nil {
		return "error"
	} else if s {
		return "engaged"
	}

	return "disengaged"
}

func (d *Door) Unlock() (<-chan error, error) {
	var err error

	ch := make(chan error, 1)

	if err = d.ifk.Outputs[*d.Lock].SetState(true); err != nil {
		return nil, err
	}

	if d.ifk.LockDelay != nil {
		time.Sleep(time.Duration(*d.ifk.LockDelay) * time.Millisecond)
	} else {
		time.Sleep(200 * time.Millisecond)
	}

	if err = d.ifk.Outputs[*d.Lock].SetState(false); err != nil {
		return nil, err
	}

	ch <- nil

	return ch, nil
}

func (d *HTTPDoor) MarshalJSON() ([]byte, error) {
	o := make(map[string]interface{})

	o["id"] = d.ID
	o["lock"] = d.LockStatus()
	o["mag"] = d.MagStatus()
	o["name"] = d.Name

	return json.Marshal(o)
}
