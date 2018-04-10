# build

```
./build.sh
```

# config

**priority**:  command line arguments > config file parameters

## use config file parameter

```
//set user1(alias)
$ pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"

//set user2(alias)
$ pi config set-credentials user2 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"

//select default user
$ pi config set-context default --user=user1


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
    user: testuser3
  name: default
current-context: default
kind: Config
preferences: {}
users:
- name: testuser2
  user:
    access-key: xxx
    region: gcp-us-central1
    secret-key: xxxxxx
- name: testuser3
  user:
    access-key: xxx
    region: gcp-us-central1
    secret-key: xxxxxx
```

## use command line arguments

```
// argument:
--server
--region
--access-key
--secret-key


//user default user
$ pi get nodes

//use specified user
$ pi --user=user2 get nodes

//specify credential
$ pi --access-key=xxx --secret-key=xxxxxx get nodes

//specify region
$ pi --region=gcp-us-central1 --access-key=xxx --secret-key=xxxxxx get nodes

//specify server
$ pi --server=https://gcp-us-central1.hypersh --access-key=xxx --secret-key=xxxxxx get nodes
```


# usage

## subcommand

```
$ pi                                                                                                                               17:12:51
pi controls the resources on Hyper GCP cluster.

Find more information at https://github.com/hyperhq/pi.

Basic Commands (Beginner):
  create      Create a resource.

Basic Commands (Intermediate):
  get         Display one or many resources
  delete      Delete resources by resources and names
  name        Name a resource

Troubleshooting and Debugging Commands:
  exec        Execute a command in a container

Other Commands:
  config      Modify piconfig files
  help        Help about any command

Usage:
  pi [flags] [options]

Use "pi <command> --help" for more information about a given command.
Use "pi options" for a list of global command-line options (applies to all commands).
```


## global options

``` 
$ pi options                                                                                                                        14:39:32
The following options can be passed to any command:

      --access-key='': AccessKey authentication to the API server
      --region='': Region of the API server
      --secret-key='': SecretKey for basic authentication to the API server
      --user='': The name of the config user to use
```

## create flags

```
$ pi create -h
Create a resource.

JSON and YAML formats are accepted.

Examples:
  # Create a pod using the data in pod.json.
  pi create -f ./pod.json

Available Commands:
  fip         Create one or more fip(s) using specified subcommand
  secret      Create a secret using specified subcommand
  service     Create a service using specified subcommand
  volume      Create a volume using specified subcommand

Options:
  -f, --filename=[]: Filename, directory, or URL to files to use to create the resource
      --validate=true: If true, use a schema to validate the input before sending it

Usage:
  pi create -f FILENAME [flags] [options]
```

### create service

```
$ pi create service -h                                                                                                                                                 16:21:58
Create a service using specified subcommand

Aliases:
service, svc

Available Commands:
  clusterip    Create a ClusterIP service.
  loadbalancer Create a LoadBalancer service.

Usage:
  pi create service [flags] [options]
```

### create secret

```
$ pi create secret -h
Create a secret using specified subcommand

Available Commands:
  docker-registry Create a secret for use with a Docker registry
  generic         Create a secret from a local file, directory or literal value

Usage:
  pi create secret [flags] [options]
```


## get flags

```
$ pi get -h
Display one or many resources

Valid resource types include:

  * nodes (aka 'no')
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

  # List a single pod in JSON output format.
  pi get -o json pod web-pod-13je7

  # List all replication controllers and services together in ps output format.
  pi get pods,services

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

## name flag

```
$ pip name -h
Name a resource(support fip only).

Examples:
  # Name a fip.
  pi name fip x.x.x.x --name=test

Available Commands:
  fip         Name a fip

Usage:
  pi name [flags] [options]
```


# example

## nodes operation example

```
$ pi get nodes
NAME              STATUS    ROLES     AGE       VERSION
gcp-us-central1   Ready     <none>    4d


$ pi get nodes --show-labels
NAME              STATUS    ROLES     AGE       VERSION   LABELS
gcp-us-central1   Ready     <none>    6d                  availabilityZone=gcp-us-central1-b|UP,defaultZone=gcp-us-central1-b,email=testuser3@test.com,resources=pod:3/20,volume:6/40,fip:0/5,service:1/5,secret:1/3,serviceClusterIPRange=10.96.0.0/12,tenantID=00a54ebcc0444bb384e48f6fd7b5597b


$ pi get nodes -o yaml
apiVersion: v1
items:
- apiVersion: v1
  kind: Node
  metadata:
    creationTimestamp: 2018-04-01T00:00:00Z
    labels:
      availabilityZone: gcp-us-central1-b|UP
      defaultZone: gcp-us-central1-b
      email: zewenzhang+gce1@gmail.com
      resources: pod:2/20,volume:1/40,fip:2/5,service:1/5,secret:1/3
      serviceClusterIPRange: 10.96.0.0/12
      tenantID: 5a91a867043f4ff18e18c4a4ed8f85a2
    name: gcp-us-central1
    namespace: ""
  spec:
    podCIDR: 10.244.0.0/16
  status:
    addresses:
    - address: 10.244.0.1
      type: Internal Gateway
    conditions:
    - lastHeartbeatTime: 2018-04-09T02:47:23Z
      lastTransitionTime: 2018-04-09T02:47:23Z
      message: kubelet is posting ready status
      reason: KubeletReady
      status: "True"
      type: Ready
  ...    
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


// get pod
$ pi get pods nginx -o yaml
apiVersion: v1
items:
- apiVersion: v1
  kind: Pod
  metadata:
    annotations:
      id: a0eabb46fccd616f4a22cbeec8547c344c3a85b7a1b95c1f88c17bb4d27b1267
      sh_hyper_instancetype: s4
      zone: gcp-us-central1-b
    creationTimestamp: 2018-04-08T14:39:52Z
    labels:
      app: nginx
    name: nginx
    namespace: default
    uid: b2aa75df-3b3a-11e8-b8a4-42010a000032
  spec:
    containers:
    - image: nginx
      name: nginx
      resources: {}
    nodeName: gcp-us-central1
  status:
    conditions:
    - lastProbeTime: null
      lastTransitionTime: 2018-04-08T14:39:52Z
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: 2018-04-08T14:39:56Z
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: 2018-04-08T14:39:52Z
      status: "True"
      type: PodScheduled
    containerStatuses:
    - containerID: hyper://2a077fec7051be72cb9435a7114c66c0d43cf66e7f6667cb34b28eeeea374a54
      image: sha256:c5c4e8fa2cf7d87545ed017b60a4b71e047e26c4ebc71eb1709d9e5289f9176f
      imageID: sha256:c5c4e8fa2cf7d87545ed017b60a4b71e047e26c4ebc71eb1709d9e5289f9176f
      lastState: {}
      name: nginx
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: 2018-04-08T14:39:56Z
    phase: Running
    podIP: 10.244.58.203
    qosClass: Burstable
    startTime: 2018-04-08T14:39:52Z
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""


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
$ pi get service
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
test-clusterip-default   ClusterIP   10.105.34.132   <none>        8080/     15s


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


// delete service
$ pi delete services test-clusterip-default
service "test-clusterip-default" deleted
```

## secret operation example

```
// create secret via yaml
$ pi create -f examples/secret/secret-dockerconfigjson.yaml
secret "test-secret-dockerconfigjson" created


// create secret via command line arguments
$ pi create secret docker-registry test-secret-gitlab --docker-server=https://registry.gitlab.com --docker-username=xjimmy --docker-password=xxx --docker-email=xjimmyshcn@gmail.com
secret "test-secret-gitlab" created


// list all secrets
$ pi get secrets
NAME                           TYPE                             DATA      AGE
test-secret-dockerconfigjson   kubernetes.io/dockerconfigjson   1         2m
test-secret-gitlab             kubernetes.io/dockerconfigjson   1         8s


// delete secret
$ pi delete secrets test-secret-gitlab
secret "test-secret-gitlab" deleted
```

## volume operation example

```
// create volume
$ pi create volume vol1 --size=1
volume vol1(1GB) created in zone gcp-us-central1-b


// list volumes
$ pi get volumes
NAME              ZONE               SIZE(GB)  CREATEDAT                  POD
test-performance  gcp-us-central1-b  50        2018-03-26T05:31:05+00:00  test-flexvolume


// get volume
$ pi get volumes test-performance -o json
[
  {
    "name": "test-performance",
    "size": 50,
    "zone": "gcp-us-central1-b",
    "pod": "test-flexvolume",
    "createdAt": "2018-03-26T05:31:05.773Z"
  }
]

// delete volume
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


$ pi get fips 35.192.x.x
FIP         NAME  CREATEDAT                  SERVICES
35.192.x.x        2018-04-08T15:27:49+00:00

$ pi get fips 35.192.x.x -o json
{
  "fip": "35.192.x.x",
  "name": "test",
  "createdAt": "2018-04-08T15:27:49.862Z",
  "services": []
}


$ pi delete fips 35.188.x.x
fip "35.188.x.x" deleted
```