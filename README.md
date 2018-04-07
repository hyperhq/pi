# build

```
./build.sh
```

# set credential

```
//set user1
$ pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"

//set user2
$ pi config set-credentials user2 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"
```

# use credential

```
$ pi --user=user1 get nodes
or
$ pi --user=user2 get nodes
or
$ pi --access-key=xxx --secret-key=xxxxxx get nodes
or
$ pi --server=https://gcp-us-central1.hypersh --access-key=xxx --secret-key=xxxxxx get nodes
or
$ pi --region=gcp-us-central1 --access-key=xxx --secret-key=xxxxxx get nodes
```

# config

**priority**:  command line arguments > config file parameters


```
// argument:
--host
--region
--access-key
--secret-key


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

# usage

## nodes operation example

```
$ pi get nodes
NAME              STATUS    ROLES     AGE       VERSION
gcp-us-central1   Ready     <none>    4d

$ pi get nodes --show-labels
NAME              STATUS    ROLES     AGE       VERSION   LABELS
gcp-us-central1   Ready     <none>    6d                  availabilityZone=gcp-us-central1-b|UP,defaultZone=gcp-us-central1-b,email=testuser3@test.com,resources=pod:3/20,volume:6/40,fip:0/5,service:1/5,secret:1/3,serviceClusterIPRange=10.96.0.0/12,tenantID=00a54ebcc0444bb384e48f6fd7b5597b
```

## pod operation example

```
// create pod via yaml
$ pi create -f examples/pod/pod-nginx.yaml
pod "nginx" created


// list pods
$ pi get pods
NAME      READY     STATUS    RESTARTS   AGE
nginx     0/1       Pending   0          11m


// get pod
$ pi get pods nginx -o yaml
apiVersion: v1
kind: Pod
metadata:
  annotations:
    id: cb318305b1ad941ea957f65503d222c0139635c28f5c2d0aeae78f853d1c72b3
    sh_hyper_instancetype: s4
    zone: gcp-us-central1-b
  creationTimestamp: 2018-04-07T13:32:57Z
  labels:
    app: nginx
  name: nginx
  namespace: default
  uid: 2f062ac9-3a68-11e8-b8a4-42010a000032
spec:
  containers:
  - image: nginx
    name: nginx
    resources: {}
  nodeName: gcp-us-central1
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: 2018-04-07T13:32:57Z
    message: '0/2 nodes are available: 1 MatchNodeSelector, 1 PodToleratesNodeTaints,
      2 NodeNotReady.'
    reason: Unschedulable
    status: "False"
    type: PodScheduled
  phase: Pending
  qosClass: Burstable


// delete pod
$ pi delete pod nginx
pod "nginx" deleted
```

## service operation example

```
// create service via yaml
$ pi create -f examples/service/service-clusterip-default.yaml
service "test-clusterip-default" created


// list services
$ pi get service
NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
test-clusterip-default   ClusterIP   10.105.151.85   <none>        8080/     10m


// get service
$ pi get service -o yaml
apiVersion: v1
items:
- apiVersion: v1
  kind: Service
  metadata:
    annotations:
      id: c6bb728fb586a8301c96b6cf7ef01c41698ccd09a1f7d469a653792d8b76227a
    creationTimestamp: 2018-04-07T13:24:52Z
    name: test-clusterip-default
    namespace: default
    uid: 0e2828f7-3a67-11e8-b8a4-42010a000032
  spec:
    clusterIP: 10.105.151.85
    ports:
    - port: 8080
      targetPort: 80
    selector:
      app: nginx
    type: ClusterIP
  status:
    loadBalancer: {}
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""


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
+---------------------+-------------------+----------+---------------------------+--------------------+
|        NAME         |       ZONE        | SIZE(GB) |         CREATEDAT         |        POD         |
+---------------------+-------------------+----------+---------------------------+--------------------+
| pt-test-performance | gcp-us-central1-b |       50 | 2018-03-26T05:31:05+00:00 | pt-test-flexvolume |
| vol1                | gcp-us-central1-b |        1 | 2018-04-07T18:26:18+00:00 |                    |
+---------------------+-------------------+----------+---------------------------+--------------------+

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

// delete volume
$ pi delete volume vol1
volume vol1 deleted

```