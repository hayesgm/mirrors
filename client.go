package main

import (
  "github.com/hayesgm/mirrors/mirrors"
  "github.com/coreos/go-etcd/etcd"
  "flag"
  "log"
)

func getEtcd() (cli *etcd.Client) {
  cli = etcd.NewClient()
  return
}

func main() {
  var domain, token string

  flag.StringVar(&domain, "domain", "", "domain to host mirror")
  flag.StringVar(&token, "token", "", "dnsimple token to manage domain")
  flag.Parse()

  if len(domain) == 0 || len(token) == 0 {
    log.Fatal("Must include domain and token")
  }

  err := mirrors.Join(getEtcd(), domain, token, "journal")
  if err != nil {
    log.Fatal(err)
  }

  hold := make(chan int)
  <- hold // Allow mirrors to run
}