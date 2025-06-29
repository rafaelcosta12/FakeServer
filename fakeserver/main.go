package main

import (
	"log"
	"os"
)

var fakeServer = FakeServer{
	DbgServerAddr: "127.0.0.1:8888",
	DbgClientAddr: "127.0.0.1:8889",
}

func main() {
	parseCommandLineArgs()
	fakeServer.Start()
}

func parseCommandLineArgs() {
	if len(os.Args) > 2 {
		fakeServer.ListenAddr = os.Args[1]
		fakeServer.RemoteAddr = os.Args[2]
		for i := 3; i < len(os.Args); i++ {
			switch os.Args[i] {
			case "-s":
				if i+1 < len(os.Args) {
					fakeServer.DbgServerAddr = os.Args[i+1]
					i++
				} else {
					log.Fatalf("Missing value for -s (dbg_server_address)")
				}
			case "-c":
				if i+1 < len(os.Args) {
					fakeServer.DbgClientAddr = os.Args[i+1]
					i++
				} else {
					log.Fatalf("Missing value for -c (dbg_client_address)")
				}
			}
		}
	} else {
		log.Fatalf("Usage: %s <listen_address> <remote_address> [-s <dbg_server_address>] [-c <dbg_client_address>]", os.Args[0])
	}
}
