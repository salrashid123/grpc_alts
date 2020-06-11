package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/alts"
	ecpb "google.golang.org/grpc/examples/features/proto/echo"
	"google.golang.org/grpc/peer"
)

var addr = flag.String("addr", "localhost:50051", "the address to connect to")
var targetServiceAccount = flag.String("targetServiceAccount", "", "the targetServiceAccount to connect to")

func callUnaryEcho(client ecpb.EchoClient, message string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var p peer.Peer

	resp, err := client.UnaryEcho(ctx, &ecpb.EchoRequest{Message: message}, grpc.Peer(&p))
	if err != nil {
		log.Fatalf("client.UnaryEcho(_) = _, %v: ", err)
	}

	ai, err := alts.AuthInfoFromPeer(&p)
	if err != nil {
		log.Fatalf("Unable to get client AuthInfoFromPeer = _, %v: ", err)
	}
	log.Printf("AuthInfo PeerServiceAccount: %v", ai.PeerServiceAccount())
	log.Printf("AuthInfo LocalServiceAccount: %v", ai.LocalServiceAccount())

	fmt.Println("UnaryEcho: ", resp.Message)
}

func main() {
	flag.Parse()

	// Create alts based credential.
	altsTC := alts.NewClientCreds(&alts.ClientOptions{
		TargetServiceAccounts: []string{*targetServiceAccount},
	})

	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(altsTC), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Make a echo client and send an RPC.
	rgc := ecpb.NewEchoClient(conn)
	callUnaryEcho(rgc, "hello world")
}
