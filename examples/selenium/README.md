# Selenium example

This repository contains a simple example of how to work with Selenium
using Python API.

To run this example, please follow [official
instructions](https://poetry.eustace.io/docs/#installation) on how to
install Poetry. After that, please prepare the environment with the
following command:

```console
$ python3 "$(command -v poetry)" install
```

This will install docker-compose and selenium. After that, please run
the docker-compose stack with following command:

```console
$ python3 "$(command -v poetry)" run docker-compose up
```

If this command will fail on absent `zyte-smartproxy-headless-proxy` image,
please build one (follow the instructions to
[zyte-smartproxy-headless-proxy](https://github.com/zytedata/zyte-smartproxy-headless-proxy)
project). If you need to provide your own API key to Zyte Smart Proxy Manager,
please check `docker-compose.yml` and update the corresponding line to
`headless-proxy` service.

Now everything is up and running. To execute the example, please do the
following:

```console
$ python "$(command -v poetry)" run ./run-example.py
```
