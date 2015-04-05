[![Build Status](http://komanda.io:8080/api/badge/github.com/mephux/envdb/status.svg?branch=master)](http://komanda.io:8080/github.com/mephux/envdb)

# Envdb - Environment Database

* NOTE: envdb is still beta software.

<img style="float:left;" height="244px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb.gif?token=AABXAYgKkzBNt0LlqD4LsRb9kpvnzp1aks5VKIX0wA%3D%3D">

Envdb turns your production, dev, cloud, etc environments into a database 
cluster you can search using [osquery](https://github.com/facebook/osquery) as the foundation.

Envdb allows you to register each computer, server or asset as a node in a cluster. Once a new
node is connected it becomes available for search from the Envdb ui.

Envdb was built using golang so the whole application, node client and server comes as one single binary.
This makes it really easy to deploy and get working in seconds.

Video Intro: [https://youtu.be/ydYr7Ykwzy8](https://youtu.be/ydYr7Ykwzy8)

# How it works.

Envdb wraps the osquery process with an agent that can communicate back to a central location.
When an agent gets a new query, it's executed and then sent back to the tcp server for rendering. Once the
request is processed it's then sent to any avaliable web clients using websockets.

Envdb has an embedded sqlite database for node storage and saved searches.

ui --websockets--> server --tcp--> node client.

## Moving Forward

I plan to add support and a plugin interface for extending what Envdb can request from a node. Currently that list of planned extentions includes: [yara](http://plusvic.github.io/yara/), [bro](https://www.bro.org/) and [memory](Volatility). The hope is to wrap these processes and query them using sql like osquery and allowing you to join on similar data points. 

Example: `select * from listening_ports a join bro_conn b on a.port = b.source_port;`

# Download

Pre-built versions of envdb are avaliable for linux 386/amd64. 
[linux downloads](https://github.com/mephux/envdb/releases)

Building on macosx is easy tho, checkout the section below.

# Building

  * `git clone https://github.com/mephux/envdb.git`
  * `cd envdb`
  * `make`

# Usage

  ```
usage: envdb [<flags>] <command> [<flags>] [<args> ...]

The Environment Database - SELECT * FROM awesome;

Flags:
  --help       Show help.
  --debug      Enable debug logging.
  --dev        Enable dev mode. (read assets from disk and enable debug
               output)
  -q, --quiet  Remove all output logging.

Commands:
  help [<command>]
    Show help for a command.

  server [<flags>]
    Start the tcp server for node connections.

  node --server=127.0.0.1 [<flags>] <node-name>
    Register a new node.
  ```

  * Server

    `envdb server`

    * Note: By default this will start the tcp server on port 3636 and the web server on port 8080.

  * Node Client

    `sudo envdb node --server <ip to server> SomeBoxName`

  * That's it - it's really that simple.

# More UI

<img style="float:left;" height="300px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-1.png?token=AABXAWJKIKgF-jy_wKmaxnhuD2snsbO0ks5VKH-fwA%3D%3D">

<img style="float:left;" height="300px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-2.png?token=AABXAcgvqnqiFViMFULsVUrfC2FWRjhwks5VKH_AwA%3D%3D">

<img style="float:left;" height="300px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-3.png?token=AABXAQeDVrKIbzu08PHKroPiltQJ6z3cks5VKH_KwA%3D%3D">

## Self-Promotion

Like envdb? Follow the repository on
[GitHub](https://github.com/mephux/envdb) and if
you would like to stalk me, follow [mephux](http://dweb.io/) on
[Twitter](http://twitter.com/mephux) and
[GitHub](https://github.com/mephux).

# TODO

  * Tests. Sorry :(
  * TLS for the agent/server communications (top of list)
  * Node/Server auth, verification and validation.
  * Code cleanup (will continue forever).
