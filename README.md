# docker-registry-quine

Naive implementation of an HTTP server compatible with the [Docker Registry V2](https://docs.docker.com/registry/spec/api/) protocol. It serves a single image called `quine:latest` which is a Docker image that runs quine itself, therefore it is a sort of Docker registry [quine](https://en.wikipedia.org/wiki/Quine_(computing)).

It was merely an exercise to understand how Docker Registry works.

## Usage

1. Bootstrap from public repo (or run binary locally)
```sh
$ docker run -it --rm -p 8080:8080 dzeromsk/quine:latest
```

```sh
Starting quine
config sha256:446a4924fb9bad1349923449a816fdab5a7c097e038e041502716eff860666ee
manifest sha256:fcf186963a5201ddf4975feab2e4ae80ebc89d1fd1354f4d3b34df7ecb4cf546
layer sha256:e0ff59e1d28d05b4c10204d5024b30676eaca82e0566c41707980f36b5283640
```

2. Run quine
```sh
docker run -it --rm -p 8081:8080 127.0.0.1:8080/quine:latest
```
```sh
Unable to find image '127.0.0.1:8080/quine:latest' locally
latest: Pulling from quine
e0ff59e1d28d: Pull complete 
Digest: sha256:fcf186963a5201ddf4975feab2e4ae80ebc89d1fd1354f4d3b34df7ecb4cf546
Status: Downloaded newer image for 127.0.0.1:8080/quine:latest
Starting quine
config sha256:bb0246a65de67d9d4db442a47a38cea21c0956345b20768e2eb4aa36d6e69173
manifest sha256:eec6b071bbc0304716acf4c5d14984b4dc322a5c3b46038bbfc6b59027d30dbc
layer sha256:e0ff59e1d28d05b4c10204d5024b30676eaca82e0566c41707980f36b5283640
```

3. And again pull quine from quine

```sh
docker pull 127.0.0.1:8081/quine:latest
```

```sh
latest: Pulling from quine
e0ff59e1d28d: Already exists 
Digest: sha256:eec6b071bbc0304716acf4c5d14984b4dc322a5c3b46038bbfc6b59027d30dbc
Status: Downloaded newer image for 127.0.0.1:8081/quine:latest
127.0.0.1:8081/quine:latest
```


## Under the hood

### Manifest

```sh
curl -s  http://127.0.0.1:8080/v2/quine/manifests/sha256:e9697b7f0f49701235ecab7273c1007c6a1d1135948b5f3cf088dfe460f1891b | jq .
```

```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "size": 344,
    "digest": "sha256:143815c10f575724f79b17084663cf09c51274a0d9c87521945151c390fb779e"
  },
  "layers": [
    {
      "mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
      "size": 4593679,
      "digest": "sha256:df63d08aae6cc3744a797b1f9d28fb7e1f942f381f88c2e5266533140d0f995c"
    }
  ]
}
```

### Config

```sh
curl -s http://127.0.0.1:8080/v2/quine/blobs/sha256:143815c10f575724f79b17084663cf09c51274a0d9c87521945151c390fb779e | jq .
```
```json
{
  "created": "2023-07-19T18:11:54.857104675+02:00",
  "author": "quine",
  "architecture": "amd64",
  "os": "linux",
  "config": {
    "Cmd": [
      "/quine"
    ],
    "Env": [
      "PATH=/"
    ]
  },
  "rootfs": {
    "diff_ids": [
      "sha256:31acb9f1b0422f08f1dcfdd80689211084ab87c0d090e269a4c6b8aa07636c62"
    ],
    "type": "layers"
  },
  "history": [
    {
      "created": "2023-07-19T18:11:54.857104675+02:00",
      "created_by": "quine"
    }
  ]
}
```

### Layer

```sh
curl -s http://127.0.0.1:8080/v2/quine/blobs/sha256:df63d08aae6cc3744a797b1f9d28fb7e1f942f381f88c2e5266533140d0f995c | tar tzv
```
```json
-rwxr-xr-x 0/0         8032079 1970-01-01 01:00 quine
```

## Related links
1. https://github.com/moby/moby/blob/master/image/spec/v1.2.md#image-json-description
1. https://en.wikipedia.org/wiki/Quine_(computing)
1. https://grahamc.com/blog/nix-and-layered-docker-images/
1. https://gist.github.com/tazjin/08f3d37073b3590aacac424303e6f745
1. https://github.com/tazjin/quinistry
4. https://tazj.in/blog/nixery-layers 
