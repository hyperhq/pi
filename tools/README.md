# upload binary to latest

## build

```
./build.sh
```

## package

```
//package for mac
./package.sh darwin amd64

//package for linux
./package.sh linux amd64
```

## release

> upload binary as assets of a release on github

```
export GITHUB_API_TOKEN=xxxxx

git tag alpha-0.1
git push origin alpha-0.1

//publish release on github (https://github.com/hyperhq/pi/releases/)


//release pi for mac
./release.sh owner=hyperhq repo=pi tag=alpha-0.1 os=darwin arch=amd64

//release pi for linux
./release.sh owner=hyperhq repo=pi tag=alpha-0.1 os=linux arch=amd64 
```