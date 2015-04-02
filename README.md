# Envdb

Envdb turns your production, dev, etc environments into a database cluster you can search across using [osquery](https://github.com/facebook/osquery).

Envdb allows you to register each computer, server or asset as a node in your cluster. Once a new
node is connected it because available for search.

Envdb was built using golang so the whole application, agent and server comes as one single binary.

# How it works.

Envdb wraps the osquery process with an agent that can communicate back to a central location using websockets.
When an agent gets a new query, it's executed and then sent back to the server for rendering.

  * transport - websockets.
  * includes an http server for rendering the UI. This also uses websockets to send data from the tcp
  server to the browser.
  * sqlite3 for storage, auth etc.

# Building

  * `git clone https://github.com/mephux/envdb.git`
  * `cd envdb`
  * `make`

# Usage

  * Server

    `envdb --debug server`

  * Agent

    `sudo envdb agent --name SomeBox --server <ip addr to server>`

  * That's it - it's really that simple.

## Self-Promotion

Like Komanda? Follow the repository on
[GitHub](https://github.com/mephux/komanda) and if
you would like to stalk me, follow [mephux](http://dweb.io/) on
[Twitter](http://twitter.com/mephux) and
[GitHub](https://github.com/mephux).

# TODO

  * TLS for the agent/server communications (top of list)
  * Agent/Server auth, verification and validation.
  * Code cleanup.
