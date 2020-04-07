package main

import (
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/kelseyhightower/envconfig"
)

type param struct {
	Worker int `default:"10"`
	Count  int `default:"100"`
}

func main() {
	var p param
	err := envconfig.Process("", &p)
	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	nameCh := make(chan string, 100)
	resCh := make(chan result, 100)
	for i := 0; i < p.Worker; i++ {
		wg.Add(1)
		go func() {
			lookupWorker(nameCh, resCh)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(resCh)
	}()
	go lineRead(nameCh, os.Stdin, p.Count)
	for res := range resCh {
		if res.Err != nil {
			log.Printf("Erorr Name:%s, IP:%s, err:%s", res.Name, res.IP, res.Err)
			continue
		}
		log.Printf("Name:%s, IP:%s", res.Name, res.IP)
	}
}

type result struct {
	Name string
	IP   string
	Err  error
}

func lookup(name string) (string, error) {
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return "", err
	}
	return addr.String(), err
}

func lookup2(name string) (string, error) {
	addrs, err := net.LookupHost(name)
	if err != nil {
		return "", err
	}
	for _, s := range addrs {
		return s, nil
	}
	return "", nil
}

func lookupWorker(nameCh <-chan string, resCh chan<- result) {
	for {
		name, ok := <-nameCh
		if !ok {
			return
		}
		addr, err := lookup2(name)
		resCh <- result{name, addr, err}
	}
}

func lineRead(nameCh chan<- string, f io.Reader, count int) {
	lines := []string{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	for i := 0; i < count; i++ {
		nameCh <- lines[i%len(lines)]
	}
	close(nameCh)
}
