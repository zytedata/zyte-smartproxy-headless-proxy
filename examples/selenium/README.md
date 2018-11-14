# Selenium example

This repository contains a simple example of how to work with Selenium
using Python API.

To run this example, please follow [official
instructions](https://pipenv.readthedocs.io/en/latest/install/) on how
to install pipenv. After that, please prepare the environment with the
following command:

```console
$ pipenv sync
```

This will install docker-compose and selenium. After that, please run
the docker-compose stack with following command:

```console
$ pipenv run docker-compose up
```

If this command will fail on absent `crawlera-headless-proxy` image,
please build one (follow the instructions to
[crawlera-headless-proxy](https://github.com/scrapinghub/crawlera-headless-proxy)
project). If you need to provide your own API key to Crawlera, please
check `docker-compose.yml` and update the corresponding line to
`headless-proxy` service.

Now everything is up and running. To execute the example, please do the
following:

```console
$ pipenv run ./run-example.py
```
