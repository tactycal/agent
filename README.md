# Tactycal ![alt "build status"](https://travis-ci.org/tactycal/agent.svg?branch=master "build status")

Tactycal agent is a tool for collecting details about the machine and a list of installed packages. The data is then submitted to Tactycal API where it's matched with a list of known issues and vulnerabilities.

## Installation

### Prebuilt binaries

Please visit ["Download agent"](https://beta.tactycal.com/agents) section in Tactycal for detailed instructions on how to install the agent on your system.

### Building from source

* [GNU Make](https://www.gnu.org/software/make/)
* [Golang (1.7+)](https://www.golang.org/)
* [Docker](https://www.docker.com/)
* [Docker Compose](https://docs.docker.com/compose/)

Execute following to get the source:

```
$ go get github.com/tactycal/agent          # checkout the code
$ cd $GOPATH/src/github.com/tactycal/agent  # move to the directory
$ make                                      # list all supported targets
```

The most important targets:

```
$ make test                      # runs unit tests
$ make build                     # builds agent 
$ make up                        # starts agents in all distributions
$ make up/<distribution>         # starts agent for specific distribution
```

Currently supported distributions are:

* centos
* debian
* rhel
* ubuntu
* openSUSE
* sles
* amzn (Amazon Linux AMI)

Run `make` or `make help` for a list of all supported targets.

## Configuration

Agent can be configured with a simple key/value configuration file. It will take `/etc/tactycal/agent.conf` as a default, unless a custom file is provided:

```
$ tactycal -f your_config.conf
```

Following configuration options are available:

| Option    | Required | Description |
|:----------|:---------|:------------|
| `token`   | Yes      | token used for authentication and authorization |
| `uri`     | No       | full API endpoint, defaults to `https://api.tactycal.com/v1` (should be used for development only) |
| `labels`  | No       | list of comma separated values that will be stored together with your host (you can also use environment variables, ex: `$SERVER_ROLE`) |
| `proxy`   | No       | URL of the proxy server |
| `timeout` | No       | set timeout for calls to Tactycal's API (check [Go's documentation](https://golang.org/pkg/time/#ParseDuration) for notation) |
| `state`   | No       | path to file where client's authentication state will be stored |

## Running the agent

Additional command line arguments can be set when running Tactycal:

| Argument      | Default                    | Description |
|:--------------|:---------------------------|:------------|
| `-f string`   | `/etc/tactycal/agent.conf` | use a configuration file |
| `-d`          |  `false`                   | output debug information |
| `-s string`   | `/var/opt/tactycal/state`  | path to where Tactycal can write its state |
| `-t duration` |  `3s`                      | client timeout for request in seconds (check [Go's documentation](https://golang.org/pkg/time/#ParseDuration) for notation) |
| `-v`          |                            | print version and exit |
| `-l`          | `false`                    | print host information and installed packages to standard output as `json` string and exit |
