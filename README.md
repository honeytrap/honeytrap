# Honeytrap [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/honeytrap/honeytrap?utm_source=badge&utm_medium=badge&utm_campaign=&utm_campaign=pr-badge&utm_content=badge) [![Go Report Card](https://goreportcard.com/badge/honeytrap/honeytrap)](https://goreportcard.com/report/honeytrap/honeytrap) [![Build Status](https://travis-ci.org/honeytrap/honeytrap.svg?branch=master)](https://travis-ci.org/honeytrap/honeytrap)

## Installation from source

```
apt install -y libpcap-dev lxc-dev

cd /usr/local
wget https://storage.googleapis.com/golang/go1.8rc3.linux-amd64.tar.gz
tar vxf go1.8rc3.linux-amd64.tar.gz

mkdir /opt/honeytrap
cd /opt/honeytrap/

export GOPATH=/opt/honeytrap
export PATH=$PATH:/usr/local/go/bin/

go get github.com/honeytrap/honeytrap/...

cp config.toml.sample config.toml
$GOPATH/bin/honeytrap

```

```
# create container base image
$ lxc-create -t download -n honeytrap -- --dist ubuntu --release trusty --arch amd64
```
## Contribute

Contributions are welcome.

### Setup your Honeytrap Github Repository

Fork Honeytrap upstream source repository to your own personal repository. Copy the URL for marija from your personal github repo (you will need it for the git clone command below).

```sh
$ mkdir -p $GOPATH/src/github.com/honeytrap/honeytrap
$ cd $GOPATH/src/github.com/honeytrap/honeytrap
$ git clone <paste saved URL for personal forked honeytrap repo>
$ cd honeytrap/honeytrap
```

###  Developer Guidelines
``Honeytrap`` community welcomes your contribution. To make the process as seamless as possible, we ask for the following:
* Go ahead and fork the project and make your changes. We encourage pull requests to discuss code changes.
    - Fork it
    - Create your feature branch (git checkout -b my-new-feature)
    - Commit your changes (git commit -am 'Add some feature')
    - Push to the branch (git push origin my-new-feature)
    - Create new Pull Request

* If you have additional dependencies for ``Honeytrap``, ``Honeytrap`` manages its dependencies using [govendor](https://github.com/kardianos/govendor)
    - Run `go get foo/bar`
    - Edit your code to import foo/bar
    - Run `make pkg-add PKG=foo/bar` from top-level directory

* If you have dependencies for ``Honeytrap`` which needs to be removed
    - Edit your code to not import foo/bar
    - Run `make pkg-remove PKG=foo/bar` from top-level directory

* When you're ready to create a pull request, be sure to:
    - Have test cases for the new code. If you have questions about how to do it, please ask in your pull request.
    - Run `make verifiers`
    - Squash your commits into a single commit. `git rebase -i`. It's okay to force update your pull request.
    - Make sure `go test -race ./...` and `go build` completes.

* Read [Effective Go](https://github.com/golang/go/wiki/CodeReviewComments) article from Golang project
    - `Honeytrap` project is fully conformant with Golang style
    - if you happen to observe offending code, please feel free to send a pull request

## Creators

**Remco Verhoef**
- <https://twitter.com/remco_verhoef>
- <https://twitter.com/dutchcoders>

## Copyright and license

Code and documentation copyright 2017 Honeytrap.

Code released under [Affero General Public License](LICENSE).
