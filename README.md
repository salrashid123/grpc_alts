## gRPC ALTS HelloWorld

Simple helloworld demonstrating GCP support for [Application Layer Transport Security](https://cloud.google.com/security/encryption-in-transit/application-layer-transport-security).  You can read more about ALTS in that article (no sense in repeating it).

`ALTS` can be thought of intrinsic platform-based security which helps ensure service->service communication uses the machine's bound identity itself.

That is, the gRPC communication will utilize and transmit an encrypted message at the application layer using keys intrinsic to the peer systems involved.  This is in contrast to user-space based security (eg, auth header, mTLS with user-space certs, etc) because the system that provides the assertion of machine identity and security is provided by the platform itself.

This repo also shows a sample envoy client server using ALTS but for HTTP traffic (you could, ofcourse use gRPC just the same w/ envoy but i'll stick with HTTP here).

For more information on ALTS configuration for Envoy, see [envoy.extensions.transport_sockets.alts.v3.Alts](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/transport_sockets/alts/v3/alts.proto#extensions-transport-sockets-alts-v3-alts)

---

This article is nothing new...its just a rehash of gRPC's [ALTS helloworld](https://github.com/grpc/grpc-go/tree/master/examples/features/encryption)

The difference in this repo is that I specifically show how to setup the VMs and actually emit the service account information from the peers (which is important to see).

---

### Setup

#### Build Client/Server

```bash
go build -o bin/client client/client.go
go build -o bin/server server/server.go
```

#### Create Service accounts/VM

```bash
gcloud iam service-accounts create alts-server --display-name "ALTS Server Service Account"
gcloud iam service-accounts create alts-client --display-name "ALTS Client Service Account"

export PROJECT_ID=`gcloud config get-value core/project`
export CLIENT_SERVICE_ACCOUNT=alts-client@$PROJECT_ID.iam.gserviceaccount.com
export SERVER_SERVICE_ACCOUNT=alts-server@$PROJECT_ID.iam.gserviceaccount.com

$ gcloud  compute  instances create alts-server \
  --service-account=$SERVER_SERVICE_ACCOUNT \
  --scopes=https://www.googleapis.com/auth/userinfo.email \
  --zone us-central1-a --image-family debian-10 --image-project=debian-cloud

$ gcloud compute  instances create alts-client \
  --service-account=$CLIENT_SERVICE_ACCOUNT \
  --scopes=https://www.googleapis.com/auth/userinfo.email \
  --zone us-central1-a --image-family debian-10 --image-project=debian-cloud
```

#### Copy binaries

```bash
$ gcloud compute scp bin/client alts-client:
$ gcloud compute scp bin/server alts-server:
```

#### Run Server

```bash
$ gcloud compute scp alts-server
$ ./server 

2020/06/08 21:58:11 AuthInfo PeerServiceAccount: alts-client@mineral-minutia-820.iam.gserviceaccount.com
2020/06/08 21:58:11 AuthInfo LocalServiceAccount: alts-server@mineral-minutia-820.iam.gserviceaccount.com
```

#### Run Client

Replace the value for your `SERVER_SERVICE_ACCOUNT` below

```bash
$ gcloud compute scp alts-client
$ ./client --addr alts-server:50051 --targetServiceAccount $SERVER_SERVICE_ACCOUNT

2020/06/08 21:58:11 AuthInfo PeerServiceAccount: alts-server@mineral-minutia-820.iam.gserviceaccount.com
2020/06/08 21:58:11 AuthInfo LocalServiceAccount: alts-client@mineral-minutia-820.iam.gserviceaccount.com
UnaryEcho:  hello world
```

#### Output

Note that in the output we've identified the intrinsic service account used at each peer.  The idea here is you can use this as an applicaiton-layer signal to allow or deny the inbound request.   Note, the client peer info is available after the RPC

---

In debug mode:

```bash
$ export GRPC_GO_LOG_VERBOSITY_LEVEL=99
$ export GRPC_GO_LOG_SEVERITY_LEVEL=info
$ ./server 
INFO: 2020/06/08 22:04:17 parsed scheme: ""
INFO: 2020/06/08 22:04:17 scheme "" not registered, fallback to default scheme
INFO: 2020/06/08 22:04:17 ccResolverWrapper: sending update to cc: {[{metadata.google.internal.:8080  <nil> 0 <nil>}] <nil> <nil>}
INFO: 2020/06/08 22:04:17 ClientConn switching balancer to "pick_first"
INFO: 2020/06/08 22:04:17 Channel switches to new LB policy "pick_first"
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 blockingPicker: the picked transport is not ready, loop back to repick
INFO: 2020/06/08 22:04:17 Subchannel picks a new address "metadata.google.internal.:8080" to connect
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc0001588f0, {CONNECTING <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to READY
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc0001588f0, {READY <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to READY
2020/06/08 22:04:17 AuthInfo PeerServiceAccount: alts-client@mineral-minutia-820.iam.gserviceaccount.com
2020/06/08 22:04:17 AuthInfo LocalServiceAccount: alts-server@mineral-minutia-820.iam.gserviceaccount.com
INFO: 2020/06/08 22:04:17 transport: loopyWriter.run returning. connection error: desc = "transport is closing"
```

```
$ ./client --addr alts-server:50051 --targetServiceAccount alts-server@mineral-minutia-820.iam.gserviceaccount.com
INFO: 2020/06/08 22:04:17 parsed scheme: ""
INFO: 2020/06/08 22:04:17 scheme "" not registered, fallback to default scheme
INFO: 2020/06/08 22:04:17 ccResolverWrapper: sending update to cc: {[{alts-server:50051  <nil> 0 <nil>}] <nil> <nil>}
INFO: 2020/06/08 22:04:17 ClientConn switching balancer to "pick_first"
INFO: 2020/06/08 22:04:17 Channel switches to new LB policy "pick_first"
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 Subchannel picks a new address "alts-server:50051" to connect
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc0001568e0, {CONNECTING <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 parsed scheme: ""
INFO: 2020/06/08 22:04:17 scheme "" not registered, fallback to default scheme
INFO: 2020/06/08 22:04:17 ccResolverWrapper: sending update to cc: {[{metadata.google.internal.:8080  <nil> 0 <nil>}] <nil> <nil>}
INFO: 2020/06/08 22:04:17 ClientConn switching balancer to "pick_first"
INFO: 2020/06/08 22:04:17 Channel switches to new LB policy "pick_first"
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 blockingPicker: the picked transport is not ready, loop back to repick
INFO: 2020/06/08 22:04:17 Subchannel picks a new address "metadata.google.internal.:8080" to connect
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc000156cb0, {CONNECTING <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to CONNECTING
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to READY
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc000156cb0, {READY <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to READY
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to READY
INFO: 2020/06/08 22:04:17 pickfirstBalancer: HandleSubConnStateChange: 0xc0001568e0, {READY <nil>}
INFO: 2020/06/08 22:04:17 Channel Connectivity change to READY
2020/06/08 22:04:17 AuthInfo PeerServiceAccount: alts-server@mineral-minutia-820.iam.gserviceaccount.com
2020/06/08 22:04:17 AuthInfo LocalServiceAccount: alts-client@mineral-minutia-820.iam.gserviceaccount.com
UnaryEcho:  hello world
INFO: 2020/06/08 22:04:17 Channel Connectivity change to SHUTDOWN
INFO: 2020/06/08 22:04:17 Subchannel Connectivity change to SHUTDOWN
```


### Envoy

The following demonstrates envoy's support for ALTS on GCP.  This snippet does _not_ use gRPC though you could adapt it to do that but instead just uses a plain HTTP upstream/downstream connection.

To use

#### Edit envoy configuration

Edit `server.yaml` and `client.yaml` and specify the upstream/downstream service accounts to use (`peer_service_accounts`).  Remember which is the peer for which end

#### Install Envoy on Client/Server

You can either run envoy within docker or (as i prefer), a direct binary.  You can get the envoy binary by "extracting" it from the docker image

On your laptop:

```bash
$ docker cp `docker create envoyproxy/envoy:v1.16.1`:/usr/local/bin/envoy .

# scp the binary to alts-server and alts-client
```


#### Copy Configuration files to client and server

Copy `server.yaml` to `alts-server` and `client-yaml` to `alts-client`

#### Run Envoy on Client/server

On `alts-server`:
```bash
./envoy -c server.yaml -l debug
```

On `alts-client`:
```bash
./envoy -c client.yaml -l debug
```


#### Access Endpoint from client

Open up a new shell on `alts-client` and run

```
curl -v http://localhost:18080/get
```

You should see the headers sent back to you from httpbin via two hops through envoy:

```
> GET /get HTTP/1.1
> Host: localhost:18080
> User-Agent: curl/7.64.0
> Accept: */*
> 
< HTTP/1.1 200 OK
< date: Tue, 09 Jun 2020 12:59:52 GMT
< content-type: application/json
< content-length: 337
< server: envoy
< access-control-allow-origin: *
< access-control-allow-credentials: true
< x-envoy-upstream-service-time: 72
< 
{
  "args": {}, 
  "headers": {
    "Accept": "*/*", 
    "Content-Length": "0", 
    "Host": "www.httpbin.org", 
    "User-Agent": "curl/7.64.0", 
    "X-Amzn-Trace-Id": "Root=1-5edf87c8-442f573c7a743fdf4419cc1d", 
    "X-Envoy-Expected-Rq-Timeout-Ms": "15000"
  }, 
  "origin": "34.70.140.15", 
  "url": "http://www.httpbin.org/get"
}
```

#### Debug logs

This is the important bit, in envoy look for the log lines on the client and server that showed `ALTS`:

* Client

```bash
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C1]   certificate_type: ALTS
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C1]   service_accont: alts-server@mineral-minutia-820.iam.gserviceaccount.com
```

* Server

```bash
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C0]   certificate_type: ALTS
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C0]   service_accont: alts-client@mineral-minutia-820.iam.gserviceaccount.com
```


The full trace of from the Client

```
[2020-06-09 12:59:52.809][3782][debug][pool] [source/common/http/http1/conn_pool.cc:95] creating a new connection
[2020-06-09 12:59:52.809][3782][debug][client] [source/common/http/codec_client.cc:34] [C1] connecting
[2020-06-09 12:59:52.809][3782][debug][connection] [source/common/network/connection_impl.cc:698] [C1] connecting to 10.128.0.11:18081
[2020-06-09 12:59:52.810][3782][debug][connection] [source/common/network/connection_impl.cc:707] [C1] connection in progress
[2020-06-09 12:59:52.810][3782][debug][pool] [source/common/http/conn_pool_base.cc:55] queueing request due to no available connections
[2020-06-09 12:59:52.811][3782][debug][connection] [source/common/network/connection_impl.cc:570] [C1] connected
[2020-06-09 12:59:52.811][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:54] [C1] TSI: doHandshake next: received: 0
[2020-06-09 12:59:52.812][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C1] TSI: doHandshake
[2020-06-09 12:59:52.814][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:79] [C1] TSI: doHandshake next done: status: 0 to_send: 1380
[2020-06-09 12:59:52.817][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C1] TSI: doHandshake
[2020-06-09 12:59:52.817][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:168] [C1] TSI: raw read result action 1 bytes 495 end_stream false
[2020-06-09 12:59:52.817][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C1] TSI: doHandshake
[2020-06-09 12:59:52.817][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:54] [C1] TSI: doHandshake next: received: 495
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:79] [C1] TSI: doHandshake next done: status: 0 to_send: 74
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:100] [C1] TSI: Handshake successful: peer properties: 3
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C1]   certificate_type: ALTS
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C1]   service_accont: alts-server@mineral-minutia-820.iam.gserviceaccount.com
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C1]   rpc_versions: 

[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:109] [C1] TSI: Handshake validation succeeded.
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:131] [C1] TSI: Handshake successful: unused_bytes: 0
[2020-06-09 12:59:52.818][3782][debug][client] [source/common/http/codec_client.cc:72] [C1] connected
[2020-06-09 12:59:52.818][3782][debug][pool] [source/common/http/http1/conn_pool.cc:244] [C1] attaching to next request
[2020-06-09 12:59:52.818][3782][debug][router] [source/common/router/router.cc:1711] [C0][S14200091602774441776] pool ready
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:216] [C1] TSI: protecting buffer size: 217
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:219] [C1] TSI: protected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:224] [C1] TSI: raw_write length 315 end_stream false
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:216] [C1] TSI: protecting buffer size: 0
[2020-06-09 12:59:52.818][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:219] [C1] TSI: protected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:216] [C1] TSI: protecting buffer size: 0
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:219] [C1] TSI: protected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:168] [C1] TSI: raw read result action 1 bytes 592 end_stream false
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:193] [C1] TSI: unprotecting buffer size: 592
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:196] [C1] TSI: unprotected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.882][3782][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:202] [C1] TSI: do read result action 1 bytes 592 end_stream false
[2020-06-09 12:59:52.882][3782][debug][router] [source/common/router/router.cc:1115] [C0][S14200091602774441776] upstream headers complete: end_stream=false
[2020-06-09 12:59:52.882][3782][debug][http] [source/common/http/conn_manager_impl.cc:1615] [C0][S14200091602774441776] encoding headers via codec (end_stream=false):
':status', '200'
'date', 'Tue, 09 Jun 2020 12:59:52 GMT'
'content-type', 'application/json'
'content-length', '337'
'server', 'envoy'
'access-control-allow-origin', '*'
'access-control-allow-credentials', 'true'
'x-envoy-upstream-service-time', '72'

[2020-06-09 12:59:52.882][3782][debug][client] [source/common/http/codec_client.cc:104] [C1] response complete
```



On the envoy server

```
[2020-06-09 12:59:52.811][2245][debug][conn_handler] [source/server/connection_handler_impl.cc:353] [C0] new connection
[2020-06-09 12:59:52.811][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C0] TSI: doHandshake
[2020-06-09 12:59:52.814][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C0] TSI: doHandshake
[2020-06-09 12:59:52.814][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:168] [C0] TSI: raw read result action 1 bytes 1380 end_stream false
[2020-06-09 12:59:52.814][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C0] TSI: doHandshake
[2020-06-09 12:59:52.814][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:54] [C0] TSI: doHandshake next: received: 1380
[2020-06-09 12:59:52.817][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:79] [C0] TSI: doHandshake next done: status: 0 to_send: 495
[2020-06-09 12:59:52.818][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C0] TSI: doHandshake
[2020-06-09 12:59:52.818][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:168] [C0] TSI: raw read result action 1 bytes 315 end_stream false
[2020-06-09 12:59:52.818][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:45] [C0] TSI: doHandshake
[2020-06-09 12:59:52.818][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:54] [C0] TSI: doHandshake next: received: 315
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:79] [C0] TSI: doHandshake next done: status: 0 to_send: 0
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:100] [C0] TSI: Handshake successful: peer properties: 3
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C0]   certificate_type: ALTS
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C0]   service_accont: alts-client@mineral-minutia-820.iam.gserviceaccount.com
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:103] [C0]   rpc_versions: 

[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:109] [C0] TSI: Handshake validation succeeded.
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:131] [C0] TSI: Handshake successful: unused_bytes: 241
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:168] [C0] TSI: raw read result action 1 bytes 0 end_stream false
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:193] [C0] TSI: unprotecting buffer size: 241
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:196] [C0] TSI: unprotected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.819][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:202] [C0] TSI: do read result action 1 bytes 241 end_stream false
[2020-06-09 12:59:52.819][2245][debug][http] [source/common/http/conn_manager_impl.cc:263] [C0] new stream
[2020-06-09 12:59:52.820][2245][debug][http] [source/common/http/conn_manager_impl.cc:731] [C0][S11516466959435954954] request headers complete (end_stream=true):
':authority', 'localhost:18080'
':path', '/get'
':method', 'GET'
'user-agent', 'curl/7.64.0'
'accept', '*/*'
'x-forwarded-proto', 'http'
'x-request-id', 'f39e4720-8c5f-43ef-8334-bc7732672c91'
'x-envoy-expected-rq-timeout-ms', '15000'
'content-length', '0'

[2020-06-09 12:59:52.820][2245][debug][http] [source/common/http/conn_manager_impl.cc:1276] [C0][S11516466959435954954] request end stream
[2020-06-09 12:59:52.820][2245][debug][router] [source/common/router/router.cc:474] [C0][S11516466959435954954] cluster 'service_httpbin' match for URL '/get'
[2020-06-09 12:59:52.820][2245][debug][router] [source/common/router/router.cc:614] [C0][S11516466959435954954] router decoding headers:
':authority', 'www.httpbin.org'
':path', '/get'
':method', 'GET'
':scheme', 'http'
'user-agent', 'curl/7.64.0'
'accept', '*/*'
'x-forwarded-proto', 'http'
'x-request-id', 'f39e4720-8c5f-43ef-8334-bc7732672c91'
'content-length', '0'
'x-envoy-expected-rq-timeout-ms', '15000'

[2020-06-09 12:59:52.820][2245][debug][pool] [source/common/http/http1/conn_pool.cc:95] creating a new connection
[2020-06-09 12:59:52.820][2245][debug][client] [source/common/http/codec_client.cc:34] [C1] connecting
[2020-06-09 12:59:52.820][2245][debug][connection] [source/common/network/connection_impl.cc:698] [C1] connecting to 34.198.151.234:80
[2020-06-09 12:59:52.820][2245][debug][connection] [source/common/network/connection_impl.cc:707] [C1] connection in progress
[2020-06-09 12:59:52.820][2245][debug][pool] [source/common/http/conn_pool_base.cc:55] queueing request due to no available connections
[2020-06-09 12:59:52.820][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:216] [C0] TSI: protecting buffer size: 0
[2020-06-09 12:59:52.820][2245][debug][connection] [source/extensions/transport_sockets/alts/tsi_socket.cc:219] [C0] TSI: protected buffer left: 0 result: TSI_OK
[2020-06-09 12:59:52.850][2245][debug][connection] [source/common/network/connection_impl.cc:570] [C1] connected
[2020-06-09 12:59:52.850][2245][debug][client] [source/common/http/codec_client.cc:72] [C1] connected
[2020-06-09 12:59:52.850][2245][debug][pool] [source/common/http/http1/conn_pool.cc:244] [C1] attaching to next request
[2020-06-09 12:59:52.850][2245][debug][router] [source/common/router/router.cc:1711] [C0][S11516466959435954954] pool ready
[2020-06-09 12:59:52.881][2245][debug][router] [source/common/router/router.cc:1115] [C0][S11516466959435954954] upstream headers complete: end_stream=false
[2020-06-09 12:59:52.881][2245][debug][http] [source/common/http/conn_manager_impl.cc:1615] [C0][S11516466959435954954] encoding headers via codec (end_stream=false):
':status', '200'
'date', 'Tue, 09 Jun 2020 12:59:52 GMT'
'content-type', 'application/json'
'content-length', '337'
'server', 'envoy'
'access-control-allow-origin', '*'
'access-control-allow-credentials', 'true'
'x-envoy-upstream-service-time', '60'

[2020-06-09 12:59:52.881][2245][debug][client] [source/common/http/codec_client.cc:104] [C1] response complete
[2020-06-09 12:59:52.882][2245][debug][pool] [source/common/http/http1/conn_pool.cc:201] [C1] response complete
```