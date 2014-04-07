gowebcache
==========

This is a cache written in go which can be accessed using curl.
Redis is required to store the data which is assumed to run on the default port (```6379```).

You can pipe data to curl using ```@-```.
Note that you should use ```--data-binary``` instead of ```--data```/```-d``` if you want to post binary data.

```
echo 'data' | curl --data-binary @- http://your_host
```

The response is a url with which you can fetch the data again.