mirrors
=======

A Go library to manage mirrored servers through etcd.  Currently hooked into DNSimple, mirrors will automatically register each server into a "POOL" record.  A single server will be an "observer" and will from time to time ping each server in the POOL.  If a server is unresponsive, it is removed from the POOL.  This is a simple way to form mirrors and load distribution.

# Installation

    import "github.com/hayesgm/mirrors"

# Usage
  
    import "github.com/coreos/go-etcd/etcd"

    mirrors.Join(etcd.NewClient(), "mydomain.com", "<DOMAIN TOKEN>")

# TODOs

* Allow other DNS servers other than DNSimple
* Don't require dependency on etcd