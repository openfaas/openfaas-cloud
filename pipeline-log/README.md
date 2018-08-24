pipeline-log function
==============================

GET - fetch pipeline log
POST - store pipeline log

Backend: S3 (tested with Minio, AWS should work)

## Example:

POST

```
curl http://192.168.0.26:31112/function/pipeline-log -d '{"repoPath": "alexellis/super-pancake", "commitSHA": "a3ef55c", "function": "slack-fn1", "source": "builder","data": "Line1\nLine2\n"}'
```

GET

```
curl "http://192.168.0.26:31112/function/pipeline-log?repoPath=alexellis/super-pancake&commitSHA=a3ef55c&function=slack-fn1" -i
```

## TBD

Add verification of sender via HMAC secret
