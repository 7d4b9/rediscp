<div align="center">
  <h1>rediscp</h1>
  <img src ="https://i.imgur.com/GAqO7LW.png" />
  <p>Copy a Redis DB easily as a file.</p>
</div>


## Motivation

To migrate Redis 4 or older to AWS Elasticache it's no possible to use Redis RDB file it's the **CLOUD**. And it's no possible top use RBD format (DUMP/RESTORE commands) due to Redis API version supported by AWS Elasticache.

## Disclaimer

Not all Redis types are supported.

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
