# opentracing-demo

- [x] You need to have docker/docker-compose
- [x]  You need to have a working local tyk developemnt environment

```
git clone  git@github.com:TykTechnologies/opentracing-demo.git $GOPATH/src/github.com/TykTechnologies/opentracing-demo
```

```
cd $GOPATH/src/github.com/TykTechnologies/opentracing-demo
```

```
make up
```

Then try registering the services. This is done separately because it might take a while before the gateways are ready to receive any calls

```
make services
```

Now, you can call ping service

```
make ping
```

You should see something linke this

```
GET /echo HTTP/1.1
Host: trace_1:6666
Accept-Encoding: gzip
Uber-Trace-Id: 4ef25064bae16148:7d58538b61abe8ea:385513c4309ded62:1
User-Agent: Go-http-client/1.1
X-Forwarded-For: 192.168.16.4
```

You can now visualize the trace on jaeger `http://localhost:16686/search`

