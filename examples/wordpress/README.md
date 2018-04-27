# Wordpress example

This is a full example using full Pi's feature, including FIP, Volume, Service, Pod, Secret.

## Create a volume


```
$ pi create volume mysql-data --size=10
volume/mysql-data

$ pi create volume wp-data --size=10
volume/wp-data

$ pi get volumes
NAME        ZONE               SIZE(GB)  CREATEDAT                  POD
mysql-data  gcp-us-central1-c  10        2018-04-26T07:04:14+00:00
wp-data     gcp-us-central1-c  10        2018-04-27T07:43:52+00:00
```

## Create a fip

```
$ pi create fip
fip/35.184.xxx.xxx

$ pi get fip
FIP             NAME  CREATEDAT                  SERVICES
35.184.xxx.xxx        2018-04-25T04:45:56+00:00

```


## Replace fip/MySQL password in yaml files

Firstly you need install `envsubst`, or you can replace envs manaually.

```
$ brew install gettext
$ echo 'export PATH="/usr/local/opt/gettext/bin:$PATH"' >> ~/.bash_profile
$ . ~/.bash_profile
```

Then replace envs to generate yaml files.

```
$ MYSQL_ROOT_PASSWORD=`echo -n "abcd1234" | base64` envsubst < mysql-secret.tpl.yaml > mysql-secret.yaml

$ FIP=35.184.xxx.xxx envsubst < wordpress-service.tpl.yaml > wordpress-service.yaml
```

## Create services/pods/secrets

```
$ pi create -f mysql-secret.yaml -f mysql-pod.yaml -f mysql-service.yaml -f wordpress-pod.yaml -f wordpress-service.yaml

$ pi get pods,services,secrets
```


## Cleanup

```
$ pi delete secret/mysql-password service/mysql service/wordpress pod/mysql pod/wordpress

$ pi delete fip 35.184.xxx.xxx
$ pi delete volume mysql-data
```