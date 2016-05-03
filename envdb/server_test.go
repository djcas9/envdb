package envdb

import "testing"

func TestServer_Server(t *testing.T) {
	if testServer.Socket.Addr() != "[::]:3636" {
		t.Fatal("TCP Server is not listening.")
	}
}

func TestServer_Server_Nodes(t *testing.T) {

	if len(testServer.Nodes) != 1 {
		t.Fatal("TCP Server has no connected nodes.")
	}

	node, err := testServer.GetNodeById(testNode.Id)

	if err != nil {
		t.Fatal("Couldn't find node in server Nodes map")
	}

	if node.Name != "test" {
		t.Fatal("Connected node has the wrong name")
	}

	if !node.Online {
		t.Fatal("Node is set to offline")
	}

	if !testServer.Alive(node.Id) {
		t.Fatal("Node is not properly connected to server.")
	}
}
