package main

import "github.com/8run0/sitechecker/internal/handler"

func main() {
	sites := []string{"https://www.medium.com/", "https://www.practical-go-lessons.com/", "https://www.stackoverflow.com/", "https://www.skysports.com", "https://www.mail.com", "https://www.guardian.com", "https://www.google.com", "https://www.kll.la", "https://www.bbc.co.uk", "https://yoox.com"}
	lrh := handler.NewLatencyRequestHandler()
	lrh.AsyncLatencyCheckSiteList(sites)
}
