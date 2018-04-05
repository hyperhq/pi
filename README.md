# build

```
./build.sh
```

# set credential

```
$ pi config set-credentials user1 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"
$ pi config set-credentials user2 --region=gcp-us-central1 --access-key="xxx" --secret-key="xxxxxx"
```

# use credential

```
$ pi --user=user1 get nodes
$ pi --user=user2 get nodes
$ pi --access-key=xxx --secret-key=xxxxxx get nodes
$ pi --server=https://gcp-us-central1.hypersh --access-key=xxx --secret-key=xxxxxx get nodes
$ pi --region=gcp-us-central1 --access-key=xxx --secret-key=xxxxxx get nodes
```

# config

>priority:  command line argument => config file

```
// argument:
--host
--region
--access-key
--secret-key

// config file:
~/.pi/config
```

# usage

```
$ pi get nodes
NAME              STATUS    ROLES     AGE       VERSION
gcp-us-central1   Ready     <none>    4d

$ pi get pods
NAME                 READY     STATUS    RESTARTS   AGE
test-flexvolume      1/1       Running   0          8d
test-service-nginx   1/1       Running   1          10d

$ pi get services
NAME                  TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
test-clusterip-none   ClusterIP   None         <none>        8080/     9d

$ pi get secrets
NAME              TYPE                             DATA      AGE
mysecret-gitlab   kubernetes.io/dockerconfigjson   1         12d
```