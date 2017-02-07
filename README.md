# Tactycal

@todo abstract

## Installation

### Prebuilt binaries

Please visit ["Download agent"](https://beta.tactycal.com/agents) section in Tactycal for detailed instructions how to install the agent on your system.

### Building from source

As the agent needs to work on different different Linux distributions the development environment heavily relies on Docker to be able to run the code inside a valid Linux environment.

This also means it's not required to checkout the code a properly setup go environment.

Requirements:
* make
* docker
* docker-compose

```
$ git clone github.com/tactycal/agent
$ cd agent
$ make test # run unit tests for all supported environments
$ make test/redhat # run unit tests for redhat only
$ make up/centos # start the agent on centos environment
$ make build/ubuntu # builds the agent for ubuntu environemnt
```

You can run `make help` to get a list of all supported make targets.

#### Running tests locally

Additional requirements:
* go (1.7+)

If you work on mac and don't want to wait slightly longer for unit tests to be executed inside a docker container you should clone the repo inside your `$GOPATH/src` folder:

```
$ go get github.com/tactycal/agent/...
$ cd $GOPATH/src/github.com/tactycal/agent
$ make testLocal
$ make testLocal/debian
```

## Configuration

Agent can be configured with a simple key/value configuration file. It will take `/etc/tactycal/agent.conf` as a default, unless a custom file is provided:

```
$ tactycal -f your_config.conf
```

Following configuration options are available:

* `token` - required API token used for authentication and authorization
* `uri` - optional full API endpoint, defaults to `https://api.tactycal.com/v1` (should be used for development only)
* `labels` - optional list of comma separated values that will be stored together with your host (you can also use environment variables, ex: `$SERVER_ROLE`)
* `proxy` - optional URL of the proxy server
* `timeout` - set timeout for calls to Tactycal's API (check [golang's documentation](https://golang.org/pkg/time/#ParseDuration) for notation)
* `state` - path to file where client's authentication state will be stored

## Running the agent

Additional command line arguments can be set when running tactycal:

* `-f string` use a configuration file (default "/opt/tactycal/etc/agent.conf")
* `-q` output only important information
* `-s string` Path to where tactycal can write it's state (default "/opt/tactycal/var/state")
* `-t duration` client timeout for request in seconds (default 3s)
* `-v` print version and exit
