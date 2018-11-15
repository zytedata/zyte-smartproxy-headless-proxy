# Puppeteer example

This repository contains a simple example of how to work with headless
Chrome using [Puppeteer](https://pptr.dev), official JS API for headless
Chrome.

To run this example, please follow [official
instructions](https://pipenv.readthedocs.io/en/latest/install/) on how
to install pipenv. After that, please prepare the environment with the
following command:

```console
$ pipenv sync
```

Also, you need to have node.js and npm installed to be
able to use Puppeteer. Please [follow the official install
guide](https://www.npmjs.com/get-npm) to obtains both tools. After you
will have them, please install Puppeteer with the following command:

```console
$ npm install
```

Now let's run headless proxy:

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
$ npm run example
```

or

```console
$ node index.js
```
