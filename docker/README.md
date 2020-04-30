# stiff

An opinionated static file web server.

[Learn more ›](https://github.com/undecidabot/stiff/blob/master/README.md)


## How to use this image

### Using volumes

```
$ docker run -v /path/to/my-website:/public:ro undecidabot/stiff
```

### Using a Dockerfile

```
FROM stiff

COPY my-website /public
COPY stiff.json /stiff.json
```
