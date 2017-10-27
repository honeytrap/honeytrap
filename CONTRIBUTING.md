# Honeytrap Contribution Guide

``Honeytrap`` community welcomes your contribution. To make the process as seamless as possible, we recommend you read this contribution guide.

## Development Workflow

Start by forking the Honeytrap GitHub repository, make changes in a branch and then send a pull request. We encourage pull requests to discuss code changes. Here are the steps in details:

### Setup your Honeytrap Github Repository
Fork [Honeytrap upstream](https://github.com/honeytrap/honeytrap/fork) source repository to your own personal repository. Copy the URL of your Honeytrap fork (you will need it for the `git clone` command below).

```sh
$ mkdir -p $GOPATH/src/github.com/honeytrap
$ cd $GOPATH/src/github.com/honeytrap
$ git clone <paste saved URL for personal forked honeytrap repo>
$ cd honeytrap
```

### Set up git remote as ``upstream``
```sh
$ cd $GOPATH/src/github.com/honeytrap/honeytrap
$ git remote add upstream https://github.com/honeytrap/honeytrap
$ git fetch upstream
$ git merge upstream/master
...
```

### Create your feature branch
Before making code changes, make sure you create a separate branch for these changes: 

```
$ git checkout -b my-new-feature
```

### Test Honeytrap server changes
After your code changes, make sure you'll:

- add test cases for the new code. 
- squash your commits into a single commit. `git rebase -i`. It's okay to force update your pull request.
- run `go test -race ./...` and `go build` completes.

### Commit changes
After verification, commit your changes. This is a [great post](https://chris.beams.io/posts/git-commit/) on how to write useful commit messages.

```
$ git commit -am 'Add some feature'
```

### Push to the branch
Push your locally committed changes to the remote origin (your fork):
```
$ git push origin my-new-feature
```

### Create a Pull Request
Pull requests can be created via GitHub. Refer to [this document](https://help.github.com/articles/creating-a-pull-request/) for detailed steps on how to create a pull request. After a Pull Request gets peer reviewed and approved, it will be merged.

## FAQs
### How does ``Honeytrap`` manages dependencies? 
``Honeytrap`` manages its dependencies using [dep]. To add a dependency:
- Run `dep ensure`

To remove a dependency
- Edit your code to not import foo/bar
- Run `dep ensure`

### What are the coding guidelines for Honeytrap?
``Honeytrap`` is fully conformant with Golang style. Refer: [Effective Go](https://github.com/golang/go/wiki/CodeReviewComments) article from Golang project. If you observe offending code, please feel free to send a pull request.
