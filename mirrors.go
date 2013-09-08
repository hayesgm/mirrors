package mirrors

import (
  "github.com/hayesgm/dnsimple"
  "github.com/hayesgm/go-etcd-lock/daemon"
  "net/http"
  "io/ioutil"
  "log"
  "github.com/coreos/go-etcd/etcd"
  "strings"
  "time"
  "fmt"
)

func getDomainParts(domain string) (pre string, post string) {
  domainParts := strings.Split(domain, ".") // We're going to assume we have a single tld e.g. www.mirrors.com or mirrors.com, not mirrors.co.uk
  
  if len(domainParts) <= 2 { // Full domain
    pre, post = "", domain
  } else {
    pre, post = domainParts[0], strings.Join(domainParts[1:], ".")
  }

  return
}

func registerDNS(domain, token string) (err error) {
  cli := &dnsimple.Client{dnsimple.NewDomainAuth(domain, token)}
  // cli.GetRecords("gofiddler.com")
  // Simple way to get my IP
  resp, err := http.Get("http://icanhazip.com/")
  if err != nil  {
    log.Fatal("Error getting ip:",err)
  }
  
  ipBytes, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    log.Fatal("Unable to parse response:",err)
  }
  ip := string(ipBytes)
  log.Println("My ip:",ip)
  
  pre, post := getDomainParts(domain)
  err = cli.CreateRecord(post, pre, dnsimple.POOL_RECORD, ip, 10, 60) // Okay, this is a good way to register into round-robin

  return
}

func removeDNS(domain, token string, record dnsimple.Record) (err error) {
  cli := &dnsimple.Client{dnsimple.NewDomainAuth(domain, token)}
  pre, post := getDomainParts(domain)
  
  err = cli.DeleteRecord(post, pre, record)

  return
}

func getRecords(domain, token string) (records []dnsimple.RecordObj, err error) {
  cli := &dnsimple.Client{dnsimple.NewDomainAuth(domain, token)}

  pre, post := getDomainParts(domain)
  return cli.GetRecords(post, pre)
}

func Join(etcdCli *etcd.Client, domain, token string) (err error) {
  err = registerDNS(domain, token)
  if err != nil {
    return
  }

  observer := func(stopCh chan int) {
    for run := true; run; {
      select {
      case <-stopCh:
        run = false // We're going to exit
      default:
        log.Println("Checking mirrors...")
        
        records, err := getRecords(domain, token)
        if err != nil {
          log.Println("Failed to get records:", err)
        } else {
          for _, v := range records {
            if v.Record.RecordType == dnsimple.POOL_RECORD {
              url := fmt.Sprintf("http://%s", v.Record.Content)
              log.Println("Testing mirror:", url)
              _, err := http.Get(url)
              if err != nil {
                log.Println("Removing:", url)
                err = removeDNS(domain, token, v.Record)
                if err != nil {
                  log.Println("Failed to remove record:", err)
                }
              } else {
                log.Println("\tSuccess")
              }
            }
          }
        }

        time.Sleep(5*time.Second)
      }
    }
  }

  daemon.RunOne(etcdCli, "observer", observer, 20)

  return
}