<div align="center">
  <h1>rediscp</h1>
  <img src ="https://i.imgur.com/GAqO7LW.png" />
  <p>Copy a Redis DB easily as a file.</p>
</div>


## Motivation

Migrate Redis data to AWS Elasticache.

Currently it's no possible to use Redis _RDB file_ because the manager service does not provide filesystem access. Also it's no possible to use _RBD format_ (`DUMP`/`RESTORE` commands) due to Redis API version supported by AWS Elasticache.

## Disclaimer

When Redis type is not supported, a _warn log_ will be printed.

Supported Redis types:
- String
- List
- Hash
- Set
- ZSet

## Usage

```sh
$> rediscp --src redis://:6379/3 --dest redis://:6380/7
```

Or you can use docker image:

```sh
$> docker run jobteaser/rediscp:v1.0.0 /rediscp --src redis://:6379/3 --dest redis://:6380/7
```

## Contributing

Your are welcome to open pull request for improve the lib.
