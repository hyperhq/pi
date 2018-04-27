```
pi create fip

pi create volume nginx-data --size=1 --zone=$ZONE

export FIP=x.x.x.x
./convert.sh


pi create -f nginx-all-in-one.yaml
```
