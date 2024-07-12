package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/raziel-aleman/hardware-monitor/internal/hardware"
	"nhooyr.io/websocket"
)

type server struct {
	subscriberMessageBuffer int
	subscribers             map[*subscriber]struct{}
	mux                     http.ServeMux
	subscribersMutex        sync.Mutex
}

type subscriber struct {
	msgs chan []byte
}

func NewServer() *server {
	s := &server{
		subscriberMessageBuffer: 10,
		subscribers:             make(map[*subscriber]struct{}),
	}
	s.mux.Handle("/", http.FileServer(http.Dir("./htmx")))
	s.mux.HandleFunc("/ws", s.subscribeHandler)
	return s
}

func (s *server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := s.subscribe(r.Context(), w, r)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (s *server) addSubscriber(subscriber *subscriber) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	s.subscribers[subscriber] = struct{}{}
	fmt.Println("Added subscriber", subscriber)
}

func (s *server) removeSubscriber(subscriber *subscriber) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	delete(s.subscribers, subscriber)
	fmt.Println("Removed subscriber", subscriber)
}

func (s *server) subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var c *websocket.Conn
	subscriber := &subscriber{
		msgs: make(chan []byte, s.subscriberMessageBuffer),
	}
	s.addSubscriber(subscriber)

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()

	ctx = c.CloseRead(ctx)
	for {
		select {
		case msg := <-subscriber.msgs:
			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()
			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				return err
			}
		case <-ctx.Done():
			s.removeSubscriber(subscriber)
			return ctx.Err()
		}
	}
}

func (s *server) broadcast(msg []byte) {
	s.subscribersMutex.Lock()
	defer s.subscribersMutex.Unlock()
	for subscriber := range s.subscribers {
		subscriber.msgs <- msg
	}
}

func main() {
	fmt.Println("Starting system monitor ...")
	srv := NewServer()

	go func(s *server) {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println("systemSection", err)
				continue
			}

			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println("diskSection", err)
				continue
			}

			cpuSection, err := hardware.GetCpuSection()
			if err != nil {
				fmt.Println("cpuSection", err)
				continue
			}

			timeStamp := time.Now().Format("2006-01-02 15:04:05")

			html := `
					<div hx-swap-oob="innerHMTL:#update-timestamp"><i class="fa-solid fa-link"></i> ` + timeStamp + `</div>
					<div hx-swap-oob="innerHMTL:#system-data">` + systemSection + `</div>
					<div hx-swap-oob="innerHMTL:#disk-data">` + diskSection + `</div>
					<div hx-swap-oob="innerHMTL:#cpu-data">` + cpuSection + `</div>
				`

			s.broadcast([]byte(html))

			time.Sleep(2 * time.Second)
		}
	}(srv)

	err := http.ListenAndServe(":8080", &srv.mux)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
