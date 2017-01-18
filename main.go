package main

import (
	"github.com/tatsushid/go-fastping"
	"net"
	"os"
	"fmt"
	"time"
	"log"
	"io"
	"os/exec"
	"strconv"
	"bufio"
	"sync"
)

var wg sync.WaitGroup

func main()  {
	l, err := os.OpenFile("38popingaev.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer l.Close()

	ml := io.MultiWriter(l, os.Stdout)

	log.SetFlags(3)
	log.SetOutput(ml)

	if len(os.Args) == 1 {
		fmt.Println(`38poPINGaev v0.2 (C)Supme 2017
Example:
	38popingaev 192.168.1.1 8.8.8.8 127.0.0.1
	38popingaev start 192.168.1.1 8.8.8.8 127.0.0.1
	38popingaev stop`)
		os.Exit(0)
	}
	if os.Args[1] == "start" {
		if len(os.Args) < 3 {
			fmt.Println("Add ip address")
			os.Exit(1)
		}

		p := exec.Cmd{
			Path: os.Args[0],
			Args: []string{os.Args[0]},
		}
		for  _, ip := range os.Args[2:] {
			p.Args = append(p.Args, ip)
		}

		err = p.Start()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		log.Printf("Start 38popingaev demon %v\n", p.Process.Pid)
		file, err := os.Create("pid")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()
		_, err = io.WriteString(file, strconv.Itoa(p.Process.Pid))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if os.Args[1] == "stop" {
		file, err := os.Open("pid")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		pid, _, err := reader.ReadLine()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		p, _ := strconv.Atoi(string(pid))
		process, _ := os.FindProcess(p)
		err = process.Kill()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Remove("pid")
		log.Print("Stop 38popingaev demon\n")
		os.Exit(0)
	}

	for i, arg := range os.Args {
		if i != 0 {
			ra, err := net.ResolveIPAddr("ip4:icmp", arg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			wg.Add(1)
			go pinger(ra)
		}
	}
	wg.Wait()
	os.Exit(1)
}

func pinger(ip *net.IPAddr) {
	defer wg.Done()
	var lp bool
	p := fastping.NewPinger()
	p.AddIPAddr(ip)
	var ok bool
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("IP Addr: %s receive, RTT: %v\n", addr.String(), rtt)
		ok = true
	}
	p.OnIdle = func() {
		if lp != ok {
			lp = ok
			log.Printf("IP Addr: %s %t", ip.String(), ok)
		}
		if !ok {
			fmt.Printf("IP Addr: %s ping timeout\n", ip.String())
		}
	}
	for {
		ok = false
		err := p.Run()
		if err != nil {
			log.Print(err)
			return
		}
		time.Sleep(time.Second)
	}
}