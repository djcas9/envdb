
<img style="float:left;" height="150px" width="150px" src="https://raw.githubusercontent.com/mephux/envdb/master/web/favicon.png?token=AABXASoBcBs5d1Il3UAJ9AO_B44fugr1ks5VKH7lwA%3D%3D">

## Envdb - Environment Database

Envdb turns your production, dev, etc environments into a database 
cluster you can search using [osquery](https://github.com/facebook/osquery).

Envdb allows you to register each computer, server or asset as a node in a cluster. Once a new
node is connected it becomes available for search from the Envdb ui.

Envdb was built using golang so the whole application, node client and server comes as one single binary.
This makes it really easy to deploy and get working in seconds.

# How it works.

Envdb wraps the osquery process with an agent that can communicate back to a central location.
When an agent gets a new query, it's executed and then sent back to the tcp server for rendering. Once the
request is processed it's then sent to any avaliable web clients using websockets.

Envdb has an embedded sqlite database for node storage and saved searches.

ui --websockets--> server --tcp--> node client.


# Building

  * `git clone https://github.com/mephux/envdb.git`
  * `cd envdb`
  * `make`

# Usage

  * Server

    `envdb server`

    * Note: By default this will start the tcp server on port 3636 and the web server on port 8080.

  * Node Client

    `sudo envdb node --server <ip to server> SomeBoxName`

  * That's it - it's really that simple.

# Envdb UI

<img style="float:left;" height="350px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-1.png?token=AABXAWJKIKgF-jy_wKmaxnhuD2snsbO0ks5VKH-fwA%3D%3D">

<img style="float:left;" height="350px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-2.png?token=AABXAcgvqnqiFViMFULsVUrfC2FWRjhwks5VKH_AwA%3D%3D">

<img style="float:left;" height="350px" src="https://raw.githubusercontent.com/mephux/envdb/master/data/envdb-3.png?token=AABXAQeDVrKIbzu08PHKroPiltQJ6z3cks5VKH_KwA%3D%3D">

## Self-Promotion

Like envdb? Follow the repository on
[GitHub](https://github.com/mephux/envdb) and if
you would like to stalk me, follow [mephux](http://dweb.io/) on
[Twitter](http://twitter.com/mephux) and
[GitHub](https://github.com/mephux).

# TODO

  * TLS for the agent/server communications (top of list)
  * Node/Server auth, verification and validation.
  * Code cleanup (will continue forever).
