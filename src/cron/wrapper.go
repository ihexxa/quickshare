package cron

import cronv3 "github.com/robfig/cron/v3"

type ICron interface {
	AddFun(spec string, cmd func()) error
	Start()
	Stop()
}

type MyCron struct {
	*cronv3.Cron
}

func NewMyCron() *MyCron {
	return &MyCron{
		Cron: cronv3.New(),
	}
}
