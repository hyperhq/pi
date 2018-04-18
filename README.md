# Build

```
$ make
```

# Config

## use config file parameter

```
//set user1(alias)
$ pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"

//set user2(alias, default region is gcp-us-central1)
$ pi config set-credentials user2 --access-key="yyy" --secret-key="yyyyyy"

//switch default user
$ pi config set-context default --user=user2


// config file:
$ cat ~/.pi/config
apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: https://*.hyper.sh:6443
  name: default
contexts:
- context:
    cluster: default
    namespace: default
    user: user3
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: user2
  user:
    access-key: xxx
    region: gcp-us-central1
    secret-key: xxxxxx
- name: user3
  user:
    access-key: yyy
    region: gcp-us-central1
    secret-key: yyyyyy
```

## use command line arguments

**priority**:  

> command line arguments --access-key,--secret-key,--region will cover the parameters in config file


```
// main global options:
--user
--server
--region
--access-key
--secret-key


//user default user
$ pi info

//use specified user
$ pi --user=user2 info

//use specified user and region
$ pi --user=user2 --region=gcp-us-central1 info

//use specify credential
$ pi --access-key=xxx --secret-key=xxxxxx info

//specify region
$ pi --region=gcp-us-central1 --access-key=xxx --secret-key=xxxxxx info

//specify server
$ pi --server=https://gcp-us-central1.hyper.sh --access-key=xxx --secret-key=xxxxxx info
```


# Usage

## show all subcommand

```
$ pi
pi controls the resources on Pi platform.

Find more information at https://github.com/hyperhq/pi.

Basic Commands (Beginner):
  create      Create a resource(support pod, service, secret, volume, fip)

Basic Commands (Intermediate):
  get         Display one or many resources
  delete      Delete resources by resources and names
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


## global options

``` 
$ pi options
The following options can be passed to any command:

  -e, --access-key='': AccessKey authentication to the API server
  -r, --region='': Region of the API server
  -k, --secret-key='': SecretKey for basic authentication to the API server
  -s, --server='': The address and port of the Kubernetes API server
  -u, --user='': The name of the config user to use

```

## create flags

```
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
  volume      Create a volume using specified subcommand

Options:
  -f, --filename=[]: Filename, directory, or URL to files to use to create the resource

Usage:
  pi create -f FILENAME [flags] [options]
```

### create volume flag

```
$ pi create volume -h
Create a secret using specified subcommand

Examples:
  # Create a new volume named vol1 with default size and zone
  pi create volume vol1

  # Create a new volume named vol1 with specified size
  pi create volume vol1 --size=1

  # Create a new volume named vol1 with specified size and zone
  pi create volume vol1 --size=1 --zone=gcp-us-central1

Options:
      --size='': Specify the volume size, default 10(GB), min 1, max 1024
      --zone='': The zone of volume to create

Usage:
  pi create volume NAME [--zone=string] [--size=int] [flags] [options]
```

### create fip flag

```
$ pi create fip -h
Create one or more fip(s) using specified subcommand

Aliases:
fip, fips

Examples:
  # Create one new fip
  pi create fip

  # Create two new fips
  pi create fip --count=2

Options:
  -c, --count=1: Specify the count of fip to allocate, default is 1

Usage:
  pi create fip [--count=int] [flags] [options]
```

## get flags

```
$ pi get -h
Display one or many resources

Valid resource types include:

  * pods (aka 'po')
  * secrets
  * services (aka 'svc')
  * volumes
  * fips

Examples:
  # List all pods in ps output format.
  pi get pods

  # List all pods in ps output format with more information (such as node name).
  pi get pods -o wide

  # List pods by lable
  pi get pods -l app=nginx

  # List a single pod in JSON output format.
  pi get -o json pod web-pod-13je7

  # List all replication controllers and services together in ps output format.
  pi get pods,services,secret

  # List one or more resources by their type and names.
  pi get services/nginx pods/nginx

Available Commands:
  fip         list fips or get a fip
  volume      list volumes or get a volume

Options:
      --no-headers=false: When using the default output format, don't print headers (default print headers).
  -o, --output='': Output format. One of: json|yaml|wide|name
  -l, --selector='': Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
  -a, --show-all=false: When printing, show all resources (default hide terminated pods.)
      --show-labels=false: When printing, show all labels as the last column (default hide labels column)
      --sort-by='': If non-empty, sort list types using this field specification.  The field specification is expressed
as a JSONPath expression (e.g. '{.metadata.name}'). The field in the API resource specified by this JSONPath expression
must be an integer or a string.

Usage:
  pi get [(-o|--output=)json|yaml|wide (TYPE [NAME | -l label] | TYPE/NAME ...) [flags] [options]
```

### get volume flag

```
$ pi get volume --help
List volumes or get a volume.

Aliases:
volume, volumes

Examples:
  # List volumes
  pi get volumes

  # Get a volume named vol1
  pi get volume vol1

  # Get volume in specified zone
  pi get volumes --zone=gcp-us-central1-b

  # Show volume name only
  pi get volumes -o name

Options:
  -o, --output='': Output format. One of: json|name
      --zone='': The zone of volume to get

Usage:
  pi get volume NAME [--zone=string] [flags] [options]
```

### get fip flag

```
pi get fip --help
List fips or get a fip.

Aliases:
fip, fips

Examples:
  # List fips
  pi get fips

  # Get a specified fip
  pi get fip x.x.x.x

  # Show ip only
  pi get fip -o ip

Options:
  -o, --output='': Output format. One of: json|ip

Usage:
  pi get fip IP [flags] [options]
```

## delete flag

```
$ pi delete -h
Delete resources by resources and names.
...
Examples:
  # Delete pods and services with same names "baz" and "foo"
  pi delete pod,service baz foo

  # Delete a pod with minimal delay
  pi delete pod foo --now

  # Force delete a pod on a dead node
  pi delete pod foo --grace-period=0 --force

  # Delete all pods
  pi delete pods --all

Available Commands:
  fip         Delete a fip
  volume      Delete a volume

Options:
      --all=false: Delete all resources, including uninitialized ones, in the namespace of the specified resource types.
      --force=false: Immediate deletion of some resources may result in inconsistency or data loss and requires
confirmation.
      --grace-period=-1: Period of time in seconds given to the resource to terminate gracefully. Ignored if negative.
      --ignore-not-found=false: Treat "resource not found" as a successful delete. Defaults to "true" when --all is
specified.
      --now=false: If true, resources are signaled for immediate shutdown (same as --grace-period=1).
  -o, --output='': Output mode. Use "-o name" for shorter output (resource/name).
      --timeout=0s: The length of time to wait before giving up on a delete, zero means determine a timeout from the
size of the object

Usage:
  pi delete (TYPE [(NAME | --all)]) [flags] [options]
```

### delete volume flag

```
$ pi delete volume -h
Delete a volume.

Aliases:
volume, volumes

Examples:
  # Delete a volume named vol1
  pi delete volume vol1

Usage:
  pi delete volume NAME [flags] [options]
```

### delete fip flag

```
$ pi delete fip -h
Delete a fip.

Aliases:
fip, fips

Examples:
  # Delete a fip
  pi delete fip x.x.x.x

Usage:
  pi delete fip IP [flags] [options]
```

## name flag

```
$ pi name -h
Name a resource(support fip only).

Examples:
  # Name a fip.
  pi name fip x.x.x.x --name=test

Available Commands:
  fip         Name a fip

Usage:
  pi name [flags] [options]
```


# Example

## get info

```
$ pi info
Region Info:
  Region                 gcp-us-central1
  AvailabilityZone       gcp-us-central1-b|UP
  ServiceClusterIPRange  10.96.0.0/12
Account Info:
  Email                  user3@test.com
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
  AvailabilityZone       gcp-us-central1-b|DOWN
  ServiceClusterIPRange  10.96.0.0/12
Account Info:
  Email                  bin@hyper.sh
  TenantID               b1aee2a7c28d4abebb9b17a0f2cfabd6
  DefaultZone            gcp-us-central1-b
  Resources              pod:1/1,volume:1/1001,fip:1/1001,service:1/1,secret:1/1
Version Info:
  Version                alpha-0.1
  Hash                   f544cd7a
  Build                  2018-04-13T17:19:11+0800
there is a new version: alpha-0.2
- (Pre-release) https://github.com/hyperhq/pi/releases/download/alpha-0.2/pi.darwin-amd64.zip
- (Pre-release) https://github.com/hyperhq/pi/releases/download/alpha-0.2/pi.linux-amd64.tar.gz
```


## pod operation example

```
// create pod via yaml
$ pi create -f examples/pod/pod-nginx.yaml
pod/nginx


// list pods
$ pi get pods
NAME      READY     STATUS    RESTARTS   AGE
nginx     1/1       Running   0          12s


// exec pod
$ pi exec nginx -- echo "hello world"
hello world

$ pi exec -it mysql -c mysql -- bash
root@mysql:/# uname -r
4.12.4-hyper



// get pod
$ pi get pod nginx -o yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    id: c48e79bf8214f3130f683765794bc23d2074709544c5224754b749627f8d0e18
    sh_hyper_instancetype: s4
    zone: gcp-us-central1-b
  creationTimestamp: 2018-04-18T06:11:32Z
  labels:
    app: nginx
    role: web
  name: nginx
  namespace: default
  uid: 576353ab-42cf-11e8-b8a4-42010a000032
spec:
  containers:
  - image: oveits/docker-nginx-busybox
    name: nginx
    resources: {}
  nodeName: gcp-us-central1
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: 2018-04-18T06:11:32Z
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: 2018-04-18T06:11:36Z
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: 2018-04-18T06:11:32Z
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: hyper://f148791445f45204a03e33dfdd6dc8363ab98e6484eb6a13e43c93ebc705e059
    image: sha256:f4d95172a4064702d438b3eb5adea2d792e846ab190d971818f7a7268df7f844
    imageID: sha256:f4d95172a4064702d438b3eb5adea2d792e846ab190d971818f7a7268df7f844
    lastState: {}
    name: nginx
    ready: true
    restartCount: 0
    state:
      running:
        startedAt: 2018-04-18T06:11:36Z
  phase: Running
  podIP: 10.244.209.165
  qosClass: Burstable
  startTime: 2018-04-18T06:11:32Z


// delete pod
$ pi delete pod nginx --now
pod "nginx" deleted
```

## service operation example

```
// create service via yaml
$ pi create -f examples/service/service-clusterip-default.yaml
service/test-clusterip-default


// list services
$ pi get services
NAME                          TYPE           CLUSTER-IP      LOADBALANCER-IP   PORT(S)             AGE
test-clusterip-nginx          ClusterIP      10.109.216.15   <none>            8080/TCP,8080/UDP   20m
test-loadbalancer-mysql       LoadBalancer   10.102.5.104    35.188.87.53      3306/TCP            26m


// get service
$ pi get service clusterip -o yaml
apiVersion: v1
kind: Service
metadata:
  annotations:
    id: 60de055bc7027e49bcad66fa336e6313e3acd322a0836caecda039610eaf20bd
  creationTimestamp: 2018-04-08T14:43:37Z
  name: test-clusterip-default
  namespace: default
  uid: 3871f7a5-3b3b-11e8-b8a4-42010a000032
spec:
  clusterIP: 10.105.34.132
  ports:
  - port: 8080
    targetPort: 80
  selector:
    app: nginx
  type: ClusterIP
status:
  loadBalancer: {}


//get pods via service's selector
$ pi get pods -l app=nginx
NAME      READY     STATUS    RESTARTS   AGE
nginx     1/1       Running   0          9s


// delete service
$ pi delete service test-clusterip-default
service "test-clusterip-default" deleted
```

## secret operation example

```
// create secret via yaml
$ pi create -f examples/secret/secret-dockerconfigjson.yaml
secret/test-secret-dockerconfigjson


// list all secrets
$ pi get secrets
NAME                           TYPE                             DATA      AGE
test-secret-dockerconfigjson   kubernetes.io/dockerconfigjson   1         2m
test-secret-gitlab             kubernetes.io/dockerconfigjson   1         8s


// delete secret
$ pi delete secret test-secret-gitlab
secret "test-secret-gitlab" deleted
```

## volume operation example

```
// create volume
$ pi create volume vol1 --size=50
volume/vol1

// create pod with volume
$ pi create -f examples/pod/pod-mysql-with-volume.yaml
pod/mysql

// list volumes
$ pi get volumes
NAME              ZONE               SIZE(GB)  CREATEDAT                  POD
vol1              gcp-us-central1-b  50        2018-03-26T05:31:05+00:00  mysql


// get volume
$ pi get volume vol1 -o json
[
  {
    "name": "vol1",
    "size": 50,
    "zone": "gcp-us-central1-b",
    "pod": "mysql",
    "createdAt": "2018-03-26T05:31:05.773Z"
  }
]

// delete volume
// first you need pods which using this volume
$ pi delete pod mysql
pod "mysql" deleted

$ pi delete volume vol1
volume "vol1" deleted
```

# fip operation example

```
$ pi create fip
35.192.x.x

$ pi create fip --count=2
35.188.x.x
35.189.x.x

$ pi get fips
FIP             NAME  CREATEDAT                  SERVICES
35.192.x.x            2018-04-08T15:27:49+00:00
35.188.x.x            2018-04-08T15:31:08+00:00
35.189.x.x            2018-04-08T15:31:10+00:00

$ pi name fip 35.192.x.x --name=test
fip 35.192.x.x renamed to test


// create loadBalancer service with fip
$ pi create -f examples/service/service-loadbalancer-nginx.yaml

$ pi get fip 35.192.x.x
FIP         NAME  CREATEDAT                  SERVICES
35.192.x.x        2018-04-08T15:27:49+00:00  test-loadbalancer-nginx

// delete service first
$ pi delete service test-loadbalancer-nginx

$ pi delete fip 35.188.x.x
fip "35.188.x.x" deleted
```