package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	cli "github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "Simple http server use standard library"
	app.Usage = "start a simple http server to demo graceful shutdown"
	app.Action = defaultAction
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:   "port,p",
			EnvVar: "PORT",
			Value:  8080,
			Hidden: false,
			Usage:  "the port you want your http server to listen on",
		},
	}
	if err := app.Run(os.Args); nil != err {
		panic(err)
	}
}

func defaultAction(ctx *cli.Context) error {
	port := ctx.Int("port")
	if port <= 0 {
		// default the port to 8080
		port = 8080
	}
	http.HandleFunc("/ping", ping)
	http.HandleFunc("/slow", slowOperation)
	svc := http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   6 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	// start the server in a seperate go routine
	go func() {
		defer wg.Done()
		if err := svc.ListenAndServe(); nil != err {
			if err != http.ErrServerClosed {
				fmt.Printf("err:%s \n", err)
			}
		}
	}()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	<-ch

	// wait forever
	if err := svc.Shutdown(context.Background()); nil != err {
		fmt.Printf("shutdown err:%s", err)
	}
	wg.Wait()
	fmt.Println("exit successfully")
	return nil
}

func ping(w http.ResponseWriter, req *http.Request) {
	ct := time.Now()
	fmt.Fprintf(w, "pong:%s", ct)
	fmt.Printf("ping :%s \n", ct)
}

// simulate very slow operation
func slowOperation(w http.ResponseWriter, req *http.Request) {
	time.Sleep(time.Second * 5)
	ct := time.Now()
	fmt.Fprintf(w, "slow:%s\n", ct)
	fmt.Printf("slow :%s \n", ct)
}
