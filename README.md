# macbot: Slack Bot for doing stuff with vSphere

`macbot` is a Slack bot written in Go to let the Mac infrastructure team at Travis CI do ChatOps-type things with our vSphere datacenter.

Right now, the main thing `macbot` does for us is let us move hosts between clusters to dedicate resources to building new macOS images. Eventually, it will be able to do more for us.

## Building and Running

To build and run `macbot`:

```sh
$ go build
$ export SLACK_API_TOKEN=xxxx
$ export VSPHERE_URL=https://foo:bar@baz/sdk
$ ./macbot
```

## Developing with Docker

`macbot` is containerized. Rather than building and running our your own machine, you can use `docker-compose` while developing:

```sh
$ docker-compose up --build
```

This will build a container to run `macbot`, then run it with the `-debug` flag. The `-debug` flag switches out the backend of the bot so that it will not talk to vSphere at all. Instead, it will use a fake in-process backend. This allows testing the messages of the bot without messing with the real datacenter.

## Deploying to Docker Swarm

We deploy this bot on a Linux VM in our MacStadium datacenter. We have [a small Docker Swarm stack](https://github.com/travis-ci/terraform-config/blob/master/macstadium-pod-1/macbot.yml) that will deploy the latest version of the image:

```sh
$ docker swarm init # only the first time
$ docker stack deploy -c macbot.yml macbot
```

The last command can be run again when new versions of the image are built to update the service.
