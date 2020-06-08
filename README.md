## gRPC ALTS HelloWorld

Simple helloworld a demonstrating GCP support for [Application Layer Transport Security](https://cloud.google.com/security/encryption-in-transit/application-layer-transport-security).  You can read more about ALTS in that article (no sense in repeating it).

`ALTS` can be thought of intrinsic platform-based security which helps ensure service->service communication uses the machine's bound identity itself.

That is, the gRPC communication will utilize and transmit the service account the system runs as itself.  This is in contrast to user-space based security (eg, auth header, mTLS, etc) because the system that provides the assertion of machine identity provided by the platform itself.

---

This article is nothing new...its just a rehash of gRPC's [ALTS helloworld](https://github.com/grpc/grpc-go/tree/master/examples/features/encryption)

The difference in this repo is that I specifically show how to setup the VMs and actually emit the service account information from the peers (which is important to see).

---

### Setup

#### Build Client/Server

```bash
go build -o bin/client client/main.go
go build -o bin/server server/main.go
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
  --image=debian-10-buster-v20200521 --zone us-central1-a --image-project=debian-cloud 

$ gcloud compute  instances create alts-client \
  --service-account=$CLIENT_SERVICE_ACCOUNT \
  --scopes=https://www.googleapis.com/auth/userinfo.email \
  --image=debian-10-buster-v20200521 --zone us-central1-a --image-project=debian-cloud 
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

### Output

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
