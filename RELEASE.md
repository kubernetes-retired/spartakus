# Release Process

This application is distributed as Docker-compatible container images.  The
images are tagged using `semver` identifiers, for example `v1.0.0`.

## Cutting a release

This repo has a `Makefile` which does all the heavy lifting.

### 1. Sync to master

Before you start, make sure you are synced to the current HEAD of the master
branch.  A common pattern is:

```
$ git fetch upstream

$ git rebase upstream/master
Current branch master is up to date.
```

### 2. Make sure your git repo is clean

```
$ git status
```

Make sure any new files are added, and all changes are committed.

### 3. Make sure it builds and tests pass

```
$ make clean all test
building: bin/amd64/spartakus
Running tests:
?       github.com/kubernetes-incubator/spartakus/cmd/spartakus	[no test files]
ok      github.com/kubernetes-incubator/spartakus/pkg/collector	0.033s
?       github.com/kubernetes-incubator/spartakus/pkg/database	[no test files]
?       github.com/kubernetes-incubator/spartakus/pkg/report	[no test files]
?       github.com/kubernetes-incubator/spartakus/pkg/version	[no test files]
ok      github.com/kubernetes-incubator/spartakus/pkg/volunteer	0.152s

Checking gofmt: PASS

Checking go vet: PASS
```

### 4. Tag it

Do not force-change a tag once it has been published upstream.  Always follow
`semver` for version numbers.

```
$ git tag -a v1.0.0 -m "v1.0.0"
Updated tag 'v1.0.0' (was a813ed1)
```

### 5. Build and push the image

```
$ make push
building: bin/amd64/spartakus
Sending build context to Docker daemon 124.7 MB
Step 1 : FROM alpine:3.4
 ---> 5f05d2ba9e65
Step 2 : MAINTAINER Tim Hockin <thockin@google.com>
 ---> Using cache
 ---> 9fc31ca429cf
Step 3 : ARG ARCH
 ---> Using cache
 ---> 5825caf919cf
Step 4 : ADD bin/${ARCH}/spartakus /spartakus
 ---> Using cache
 ---> 01df31f8d33c
Step 5 : RUN apk update --no-cache && apk add ca-certificates
 ---> Using cache
 ---> 9d5a07404110
Step 6 : ENTRYPOINT /spartakus
 ---> Using cache
 ---> 1b5c33110ccf
Successfully built 1b5c33110ccf
The push refers to a repository [gcr.io/google_containers/spartakus-amd64] (len: 1)
1b5c33110ccf: Image already exists
9d5a07404110: Image already exists
01df31f8d33c: Image already exists
5f05d2ba9e65: Image already exists
v1.0.0: digest: sha256:f7f137170f1624120ac716b8b9b848c7a6deb005873fb9a68e0e229dfff4531a size: 9352

pushed: gcr.io/google_containers/spartakus-amd64:v1.0.0
```

### 6. Publish the release

https://github.com/kubernetes-incubator/spartakus/releases/new
