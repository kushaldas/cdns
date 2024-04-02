A stupid DNS forwarder, by default it forwards the queries to the `systemd-resolved`.

It currently logs things into local [redis](https://redis.io/).

You can use [just](https://just.systems/man/en/) to build the project.

## How to build?

`just build`

## How to run?

Make sure `redis` server is running locally first.

The following example to start the server on port 53 locally and will connect to `1.1.1.1` for all DNS queries.

```
./cdns 

```

You can pass a different DNS resolver using the `--remote` command line option.

## LICENSE: GPL-3.0-or-later

