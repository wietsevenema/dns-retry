# DNS retry

Exploring what's needed to create a DNS forwarder. This one takes your DNS query and forwards it to other nameservers. 

If you configure more than one nameserver, they are tried one by one in order of appearance, if a request fails (timeout or nameserver unreachable), or when the server returns a Non-Existent domain response. 

## Try it
```
go run main.go -listen localhost:5300
```

```
dig @localhost -p 5300 google.com
```

## Credits
- https://github.com/skynetservices/skydns
- https://github.com/miekg/dns
