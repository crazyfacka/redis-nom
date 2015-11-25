package main

type config struct {
	Redis struct {
		Sentinel []string
	}
	PagerDuty struct {
		Key string
	}
	Nsq struct {
		Address []string
		Topic   string
	}
}

type pod struct {
	Addr       string
	Slaves     []string
	SlaveCount int
}
