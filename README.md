pi README
-------------------------

For more about pi, please see https://docs.hyper.sh/pi

<!-- TOC depthFrom:1 depthTo:6 withLinks:1 updateOnSave:1 orderedList:0 -->

- [Build](#build)
- [Config](#config)
	- [use config file parameter](#use-config-file-parameter)
	- [use command line arguments](#use-command-line-arguments)
- [Usage](#usage)
	- [show all subcommand](#show-all-subcommand)
	- [show help](#show-help)
- [Basic Example](#basic-example)
	- [get info](#get-info)
	- [check new pi version](#check-new-pi-version)
	- [create resource](#create-resource)
		- [create from file](#create-from-file)
		- [create from flag](#create-from-flag)
	- [list resources](#list-resources)
	- [get resource detail](#get-resource-detail)
	- [delete resource](#delete-resource)
- [Advance Example](#advance-example)
	- [volume operation](#volume-operation)
		- [create volume in specified zone](#create-volume-in-specified-zone)
		- [use volume in pod](#use-volume-in-pod)
	- [pod operation](#pod-operation)
		- [pod exec](#pod-exec)
		- [pod run](#pod-run)
		- [pod list](#pod-list)
		- [delete pod immediately](#delete-pod-immediately)
		- [pod in zone](#pod-in-zone)
	- [fip operation](#fip-operation)
		- [name fip](#name-fip)
		- [allocate multiple fips](#allocate-multiple-fips)
	- [service operation](#service-operation)
		- [add clusterip for pod](#add-clusterip-for-pod)
		- [add loadbalancer for pod](#add-loadbalancer-for-pod)
	- [delete all resources](#delete-all-resources)
- [Tutorials](#tutorials)
	- [Wordpress example](#wordpress-example)

<!-- /TOC -->

# Build

```
$ make
```

# Config

## use config file parameter

```
//set user1(alias)
$ pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"
User "user1" set.

//set user2(alias, default region is gcp-us-central1)
$ pi config set-credentials user2 --access-key="yyy" --secret-key="yyyyyy"
User "user2" set.

//switch default user
$ pi config set-context default --user=user2
Context "default" modified.

//delete credentials
$ pi config delete-credentials user1
deleted credentials user1 from /Users/xjimmy/.pi/config

//view config:
$ pi config view
or
$ cat ~/.pi/config
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://*.hyper.sh:443
  name: default
contexts:
- context:
    cluster: default
    namespace: default
    user: user2
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: user2
  user:
    access-key: yyy
    region: gcp-us-central1
    secret-key: yyyyyy
```

## use command line arguments

**priority**:  

> command line arguments --region, --access-key, --secret-key will cover the parameters in config file

```
//main global options
$ pi options
The following options can be passed to any command:

  -e, --access-key='': AccessKey authentication to the API server
  -r, --region='': Region of the API server
  -k, --secret-key='': SecretKey for basic authentication to the API server
  -s, --server='': The address and port of the Kubernetes API server
  -u, --user='': The name of the config user to use


//use default user
$ pi info

//use specified user
$ pi --user=user2 info

//use specified region
$ pi --region=gcp-us-central1 info

//use specified credential
$ pi --access-key=xxx --secret-key=xxxxxx info

//use specified user and region
$ pi --user=user2 --region=gcp-us-central1 info

//use specified user and server
$ pi --server=https://gcp-us-central1.hyper.sh:443 --user=user3 info
```


# Usage

## show all subcommand

```
$ pi
pi controls the resources on Pi platform.

Find more information at https://docs.hyper.sh/pi.

Basic Commands (Beginner):
  create      Create a resource(support pod, service, secret, volume, fip)

Basic Commands (Intermediate):
  get         Display one or many resources
  delete      Delete resources by resources and names
  run         Run a particular image on the cluster
  name        Name a resource

Troubleshooting and Debugging Commands:
  exec        Execute a command in a container

Other Commands:
  config      Modify pi config file
  help        Help about any command
  info        Print region and user info

Usage:
  pi [flags] [options]

Use "pi <command> --help" for more information about a given command.
Use "pi options" for a list of global command-line options (applies to all commands).
```

## show help

```
$ pi create -h
$ pi create --help
$ pi help create

// For example
$ pi create -h
Create a resource(pod, service, secret, volume, fip).

JSON and YAML formats are accepted(pod, service, secret).

Examples:
  # Create a pod using the data in yaml.
  pi create -f examples/pod/pod-nginx.yaml

  # Create multiple pods using the data in yaml.
  pi create -f pod-test1.yaml -f pod-test2.yaml

  # Create a service using the data in yaml.
  pi create -f examples/service/service-nginx.yaml

  # Create a secret using the data in yaml.
  pi create -f examples/secret/secret-dockerconfigjson.yaml

Available Commands:
  fip         Create one or more fip(s) using specified subcommand
  secret      Create a secret using specified subcommand
  service     Create a service using specified subcommand
  volume      Create a volume using specified subcommand

Options:
  -f, --filename=[]: Filename, directory, or URL to files to use to create the resource

Usage:
  pi create -f FILENAME [flags] [options]
```


# Basic Example

## get info

```
$ pi info
Region Info:
  Region                 gcp-us-central1
  AvailabilityZone       gcp-us-central1-a|UP,gcp-us-central1-c|UP
  ServiceClusterIPRange  10.96.0.0/12
Account Info:
  Email                  test@hyper.sh
  TenantID               00a54ebcc0444bb384e48f6fd7b5597b
  DefaultZone            gcp-us-central1-b
  Resources              pod:1/20,volume:1/40,fip:1/5,service:4/5,secret:1/3
Version Info:
  Version                alpha-0.1
  Hash                   0ade6742
  Build                  2018-04-13T10:16:19+0800
```

## check new pi version

```
$ pi info --check-update
Region Info:
  Region                 gcp-us-central1
  AvailabilityZone       gcp-us-central1-a|UP,gcp-us-central1-c|UP
  ServiceClusterIPRange  10.96.0.0/12
Account Info:
  Email                  test@hyper.sh
  TenantID               00a54ebcc0444bb384e48f6fd7b5597b
  DefaultZone            gcp-us-central1-b
  Resources              pod:1/20,volume:1/40,fip:1/5,service:4/5,secret:1/3
Version Info:
  Version                alpha-0.1
  Hash                   0ade6742
  Build                  2018-04-13T10:16:19+0800

There is a new version: v1.9-b18042710
- https://github.com/hyperhq/pi/releases/download/v1.9-b18042710/pi.darwin-amd64.zip
- https://github.com/hyperhq/pi/releases/download/v1.9-b18042710/pi.linux-amd64.tar.gz
```


## create resource

Supported resources:
- volume
- fip
- pod (support create from file)
- servie (support create from file)
- secret (support create from file)

### create from file

> Only pod, service, secret support create from yaml/json

create resource from yaml

```
$ pi create -f examples/pod/pod-nginx.yaml
pod/nginx-from-yaml

$ pi create -f examples/service/service-nginx.yaml
service/test-nginx

$ pi create -f examples/secret/secret-dockercfg.yaml
secret/test-secret-dockercfg
```

create resource from json

```
$ pi create -f examples/pod/pod-nginx.json
pod/nginx-from-json
```

### create from flag

```
//create volume
$ pi create volume vol1 --size=1
volume/vol1

//create fip
$ pi create fip
fip/35.202.x.x

//create pod
$ pi run my-nginx --image=nginx
pod "my-nginx" created

//create clusterip service
$ pi create service clusterip my-cs --tcp=5678:8080
service/my-cs

//create loadbalancer service
$ pi create service loadbalancer my-lbs --tcp=5678:8080 -f=35.202.x.x -l=role=web,zone=gcp-us-central1-a
service/my-lbs

//create docker-registry secret
$ pi create secret docker-registry my-secret1 \
  --docker-username=DOCKER_USER \
  --docker-password=DOCKER_PASSWORD \
  --docker-email=DOCKER_EMAIL
secret/my-secret1

//create generic secret
$ pi create secret generic my-secret2 --from-literal=key1=supersecret --from-literal=key2=topsecret
secret/my-secret2
```


## list resources

```
// list pods
$ pi get pods
NAME      READY     STATUS    RESTARTS   AGE
nginx     1/1       Running   0          12s

// list service
$ pi get services
NAME                      TYPE           CLUSTER-IP       LOADBALANCER-IP   PORT(S)             AGE
my-cs                     ClusterIP      10.104.250.99    <none>            5678/TCP            11m
my-lbs                    LoadBalancer   10.104.104.135   35.202.x.x        5678/TCP            5m

// list secret
$ pi get secrets
NAME                    TYPE                             DATA      AGE
my-secret1              kubernetes.io/dockerconfigjson   1         5m
my-secret2                                               2         4m

// list volume (it will show related pod)
$ pi get volumes
NAME  ZONE               SIZE(GB)  CREATEDAT                  POD
vol1  gcp-us-central1-a  1         2018-04-27T04:24:49+00:00  nginx

// list fip (it will show related services)
$ pi get fips
FIP             NAME  CREATEDAT                  SERVICES
35.202.x.x            2018-04-27T04:19:27+00:00  my-lbs
```

## get resource detail

get subcommand support `-o`(`--output`)
- for pod, service, secret, output format could be one of: json|yaml|wide|name
- for volume, output format could be one of: json|name
- for fip, output format could be one of: json|ip

```
// get pod detail
$ pi get pod nginx -o yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    id: 83206f7e2428e91741aff61ce1f10e2be2fb8ef415a06f42ef4aeae9f14fa43c
    sh_hyper_instancetype: s4
    zone: gcp-us-central1-a
  creationTimestamp: 2018-04-27T04:06:53Z
  labels:
    run: nginx
  name: nginx
  uid: 6b1e2cdf-49d0-11e8-8ca0-42010a7f0003
spec:
  containers:
  - image: nginx
    imagePullPolicy: IfNotPresent
    name: nginx
    resources: {}
  dnsPolicy: ClusterFirst
  nodeName: gcp-us-central1
  restartPolicy: Always
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: 2018-04-27T04:06:53Z
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: 2018-04-27T04:06:55Z
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: 2018-04-27T04:06:53Z
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: hyper://da1f9ea02379316a4ca489108346ef1dea23605a2930106e2a621bf87f445e16
    image: sha256:b175e7467d666648e836f666d762be92a56938efe16c874a73bab31be5f99a3b
    imageID: sha256:b175e7467d666648e836f666d762be92a56938efe16c874a73bab31be5f99a3b
    lastState: {}
    name: nginx
    ready: true
    restartCount: 0
    state:
      running:
        startedAt: 2018-04-27T04:06:55Z
  phase: Running
  podIP: 10.244.144.65
  qosClass: Burstable
  startTime: 2018-04-27T04:06:53Z


// get volume detail
$ pi get volumes vol1 -o json
{
  "name": "vol1",
  "size": 1,
  "zone": "gcp-us-central1-a",
  "pod": "",
  "createdAt": "2018-04-27T04:24:49.804Z"
}
```

## delete resource

```
//delete single resource
$ pi delete pod nginx
pod "nginx" deleted

//delete multiple resources of single type
$ pi delete pods nginx nginx-from-yaml
pod "nginx" deleted
pod "nginx-from-yaml" deleted

//delete all resources of single type
$ pi delete service --all
service "my-cs" deleted
service "my-lbs" deleted

//delete multiple type resources (only support pod, service and secret)
$ pi delete pods/nginx-from-json secrets/my-secret
pod "nginx-from-json" deleted
secret "my-secret" deleted
```

# Advance Example


## volume operation

### create volume in specified zone
```
//check zone info
$ pi info | grep Zone
  AvailabilityZone       gcp-us-central1-a|UP,gcp-us-central1-c|UP
  DefaultZone            gcp-us-central1-a

//create volume with --zone
$ pi create volume vol2 --size=1 --zone=gcp-us-central1-c
volume/vol2
```

### use volume in pod

> pod and volume should be in the same zone

```
//create volume first
$ pi create volume nginx-data --size=1
volume/nginx-data

//create pod with volume
$ cat examples/pod/pod-nginx-with-volume.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-with-volume
  labels:
    app: nginx
    role: web
spec:
  containers:
  - name: nginx
    image: oveits/docker-nginx-busybox
    volumeMounts:
      - name: persistent-storage
        mountPath: /data
  volumes:
    - name: persistent-storage
      flexVolume:
        options:
          volumeID: nginx-data

$ pi create -f examples/pod/pod-nginx-with-volume.yaml
pod/nginx-with-volume

//check pod
$ pi get pods nginx-with-volume -o yaml | grep volumes -A5
  volumes:
  - flexVolume:
      driver: ""
      options:
        volumeID: nginx-data
    name: persistent-storage

//check volume (volume had been associated to pod)
$ pi get volumes nginx-data
NAME        ZONE               SIZE(GB)  CREATEDAT                  POD
nginx-data  gcp-us-central1-a  1         2018-04-27T15:24:31+00:00  nginx-with-volume
```

## pod operation

### pod exec

> exec command in running pod

```
// execute command line in pod
$ pi exec nginx -- echo "hello world"
hello world

// exec pod with interactive
$ pi exec -it mysql -c mysql -- bash
root@mysql:/# uname -r
4.12.4-hyper
root@mysql:/# exit
$
```

### pod run

> run pod and execute command in container

```
//run pod
$ pi run -it nginx --limits="memory=1024Mi" --image=nginx --labels="app=nginx,env=prod"
pod "nginx" created

//run pod once
$ pi run -it --rm busybox --image=busybox -- echo hello world
hello world
pod "busybox" deleted

//run pod with interactive
$ pi run -it --rm busybox --limits="memory=64Mi" --image=busybox --restart=Never --env="MODE=dev" -- sh
/ # ls
bin   dev   etc   home  lib   proc  root  sys   tmp   usr   var
/ # exit
pod "busybox" deleted
```


### pod list

filter pods by label
```
$ pi get pods -l app=nginx
NAME              READY     STATUS    RESTARTS   AGE
nginx             1/1       Running   0          23s
```

show all pods(include stopped)
```
$ pi get pods -a
NAME              READY     STATUS      RESTARTS   AGE
nginx             1/1       Running     0          38m
test-runonce      0/1       Succeeded   0          12m
```

### delete pod immediately

```
$ pi delete pod nginx --now

$ pi delete pod nginx --grace-period=0
```

### pod in zone

To create pod in a specified zone:
- the 'zone' key in 'spec.nodeSelector' of pod must be specified
- the value of 'zone' should be a availability zone(use `pi info` to get availabilityZone)

```
$ cat examples/pod/pod-with-zone.yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-with-zone
spec:
  containers:
  - name: busybox
    image: busybox
  nodeSelector:
    zone: gcp-us-central1-a

$ pi create -f examples/pod/pod-with-zone.yaml
pod/busybox-with-zone
```

## fip operation

### name fip

> the fip name is just a remark

```
$ pi name fip 35.193.x.x --name=production
fip "35.192.x.x" named to "production"
```

### allocate multiple fips

```
$ pi create fip --count=2
fip/35.193.x.x
fip/35.192.x.x
```

## service operation

Access pod via fip from cluster:
- create a pod with label
- create a clusterip type service, -l(--selector) must be specified

### add clusterip for pod

```
//run pod with label 'app=nginx-internal'
$ pi run my-nginx-internal --image nginx -l=app=nginx-internal
pod "my-nginx-internal" created

//create clusterip service with selector 'app=nginx-internal'
$ pi create service clusterip nginx-internal --tcp=8080:80 --selector=app=nginx-internal
service/nginx-internal

//check pod status
$ pi get pods -l app=nginx-internal --show-labels
NAME                READY     STATUS    RESTARTS   AGE       LABELS
my-nginx-internal   1/1       Running   0          36s       app=nginx-internal

//check services
$ pi get services nginx-internal -o yaml | grep -E "(clusterIP|selector):" -A1
  clusterIP: 10.105.238.13
  ports:
--
  selector:
    app: nginx-internal
```

access nginx via clusterip
```
$ pi run -it --rm busybox --image=busybox -- sh
/ # wget -qO- http://10.105.238.13:8080 | grep title
<title>Welcome to nginx!</title>
/ # exit
pod "busybox" deleted
```


### add loadbalancer for pod

Access pod via fip from internet:
- allocate a fip(floating IP)
- create a pod with label
- create a loadbalancer type service, -f(--loadbalancerip) and -l(--selector) must be specified

```
//allocate fip
$ pi create fips
fip/35.193.x.x

//run pod with label 'app=nginx-external'
$ pi run my-nginx-external --image nginx -l=app=nginx-external
pod "my-nginx-external" created

//create loadbalancer service with selector 'app=nginx-external' and fip 35.193.x.x
$ pi create service loadbalancer my-nginx-external --tcp=8080:80 -f=35.193.x.x --selector=app=nginx-external
service/my-nginx-external

//check fip (fip had been related to service)
$ pi get fip 35.193.x.x
FIP             NAME  CREATEDAT                  SERVICES
35.193.x.x            2018-04-27T15:21:03+00:00  my-nginx-external

//check pod status
$ pi get pods -l app=nginx-external --show-labels
NAME                READY     STATUS    RESTARTS   AGE       LABELS
my-nginx-external   1/1       Running   0          1m        app=nginx-external

//get pod ip
$ pi get pods my-nginx-external -o yaml | grep podIP
  podIP: 10.244.29.17

//check services
$ pi get services my-nginx-external -o yaml | grep -E "(clusterIP|selector):" -A1
  clusterIP: 10.107.218.6
  loadBalancerIP: 35.192.x.x
--
  selector:
    app: nginx-external
```

access nginx via fip
```
$ pi run -it --rm busybox --image=busybox -- sh
/ # wget -qO- http://35.192.x.x:8080 | grep title             # use loadbalancerip(fip)
<title>Welcome to nginx!</title>
/ # wget -qO- http://my-nginx-external:8080 | grep title      # use service name
<title>Welcome to nginx!</title>
/ # wget -qO- http://10.107.218.6:8080 | grep title           # use clusterip
<title>Welcome to nginx!</title>
```

## delete all resources

- `service` should be deleted before delete `fip`
- `pod` should be deleted before delete `volume`

```
$ pi delete pods,services,secrets --all
pod "my-nginx-external" deleted
pod "my-nginx-internal" deleted
service "my-nginx-external" deleted
service "nginx-internal" deleted
secret "test-secret-dockercfg" deleted

$ pi delete volumes --all
volume "nginx-data" deleted
volume "vol1" deleted
volume "vol2" deleted

$ pi delete fips --all
fip "35.193.x.x" deleted
fip "35.192.x.x" deleted
```

# Tutorials

## Wordpress example

[detail](examples/wordpress/README.md)
