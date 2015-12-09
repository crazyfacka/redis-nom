package main

type config struct {
	General struct {
		Env string
	}
	Cache struct {
		Mastername string
		Sentinel   []string
	}
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
