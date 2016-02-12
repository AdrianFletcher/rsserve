RSServe - A simple static HTML server for Redis (and Ember)
-

## Binaries
Binaries for most major platforms can be found [here](http://adrianfletcher.org/posts/rsserve-a-simply-static-server.

## Building RedisStaticServer

> % go build rsserve.go

## Usage
> rsserve -c "path/to/config.json/"

## Example Configuration

```json
{
    "redis_pass": "",
    "redis_address": "localhost",
    "redis_port": 6379,

    "key_prefix": "production",
    "key_suffix": ":index.html",

    "http_address": "test.com",
    "http_port": 80
}
```