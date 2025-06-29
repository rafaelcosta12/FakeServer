package main

import (
	"io"
	"log"
	"net"
	"sync"
)

type FakeServer struct {
	ListenAddr    string
	RemoteAddr    string
	DbgServerAddr string
	DbgClientAddr string
}

func (fs *FakeServer) Start() {
	ln, err := net.Listen("tcp", fs.ListenAddr)
	if err != nil {
		log.Fatalf("[*] Failed to listen on %s: %v", fs.ListenAddr, err)
	}
	defer ln.Close()
	log.Printf("[*] Listening on %s", fs.ListenAddr)

	for {
		clientConn, err := ln.Accept()
		if err != nil {
			log.Printf("[!] Failed to accept connection: %v", err)
			return
		}
		handleClientConnection(clientConn)
	}
}

func addDebugForwarder(fwdConn net.Conn, listenAddr string, destination string) {
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("[*] Failed to listen on %s: %v", listenAddr, err)
	}
	defer ln.Close()
	log.Printf("[*] Started debug session on %s -> %s", listenAddr, fwdConn.RemoteAddr())

	for {
		debugConn, err := ln.Accept()
		if err != nil {
			log.Printf("[!] Failed to accept connection: %v", err)
			return
		}
		go func() {
			defer debugConn.Close()
			log.Printf("[*] Debug connection established %s --> %s", debugConn.RemoteAddr(), destination)
			forwardMessages(debugConn, fwdConn, "debug->"+destination, true)
		}()
	}
}

func handleClientConnection(clientConn net.Conn) {
	log.Printf("[*] Accepted connection from %s", clientConn.RemoteAddr())

	serverConn, err := net.Dial("tcp", fakeServer.RemoteAddr)
	if err != nil {
		log.Printf("[!] Failed to connect to remote server %s: %v", fakeServer.RemoteAddr, err)
		return
	}

	log.Printf("[*] Connected to remote server %s", fakeServer.RemoteAddr)
	log.Printf("[*] Starting message handlers ...")
	log.Printf("")

	go addDebugForwarder(serverConn, fakeServer.DbgServerAddr, "server")
	go addDebugForwarder(clientConn, fakeServer.DbgClientAddr, "client")
	go handleMessageFowarders(clientConn, serverConn)
}

func handleMessageFowarders(clientConn net.Conn, serverConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		forwardMessages(clientConn, serverConn, "client", false)
	}()
	go func() {
		defer wg.Done()
		forwardMessages(serverConn, clientConn, "server", false)
	}()

	wg.Wait()
	clientConn.Close()
	serverConn.Close()
}

func forwardMessages(src net.Conn, dst net.Conn, direction string, appendZero bool) {
	buf := make([]byte, 4096)
	for {
		n, err := src.Read(buf)
		if n > 0 {
			log.Printf("(%s) -> %s [%X]", direction, string(buf[:n]), buf[:n])
			_, writeErr := dst.Write(buf[:n])
			if writeErr != nil {
				log.Printf("Error writing to client: %v", writeErr)
				return
			}
			if appendZero {
				// Append a zero byte to the end of the message
				_, writeErr = dst.Write([]byte{0})
				if writeErr != nil {
					log.Printf("Error writing zero byte to client: %v", writeErr)
					return
				}
			}
		}
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading from server: %v", err)
			}
			return
		}
	}
}
