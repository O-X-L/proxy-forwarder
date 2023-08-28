# Proxy Forwarder

This tool is specifically designed to solve a problem when using proxy servers:

* There is no clean way of forwarding traffic to a remote proxy server that is outside your layer 2 network
* Some proxy servers (_like Squid_) do not support redirecting the traffic using DNAT


<a href="https://wiki.superstes.eu/en/latest/1/network/squid.html#transparent-proxy">
<img src="https://github.com/superstes/proxy-forwarder/blob/latest/docs/squid_remote.png" alt="Remote Proxy Server" width="600"/>
</a>

For more information about Squid see: [Superstes Wiki - Squid](https://wiki.superstes.eu/en/latest/1/network/squid.html)

----

## How does it work?

This tool is based on [go-gost](https://gost.run/en/tutorials/redirect/) but I stripped all unecessary features/functions/dependencies from it.

See also: [gost documentation](https://wiki.superstes.eu/en/latest/1/network/gost.html)

### Usage

```bash
  -P 'Listen port' (required)
  -F 'Proxy server to forward the traffic to' (required, Example: 'http://192.168.0.1:3128')
  -T 'Run in TProxy mode' (default: false)
  -M 'Mark to set for TProxy traffic' (default: None)
  -m 'Set a metrics service address (prometheus)' (Example: '127.0.0.1:9000', Docs: 'https://gost.run/en/tutorials/metrics/')
  -V 'Show version'
  -D 'Enable debug mode'
```

### It does

* Bind to localhost (_127.0.0.1 & ::1_) for tcp & udp
* Allow you to redirect traffic to the forwarder using:

  * Destination NAT (_default_)
  * or [TProxy Mode](https://docs.kernel.org/networking/tproxy.html) (see also: [extended documentation](https://wiki.superstes.eu/en/latest/1/network/nftables.html#tproxy))

* Forward the traffic to the server defined using the `-F` flag

----


## Build

* [Install go](https://go.dev/doc/install)
* Download this repository
* Build the binary

  ```bash
  env GOOS=linux GOARCH=amd64 bash scripts/build.sh

  # use 'CGO_ENABLED' if you see this error on the target system:
  > error "/lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32'" on target system

  env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 bash scripts/build.sh
  ```

* Copy the new binary to the target host(s)
* Run

  ```bash
  proxy_forwarder -P 4138 -F http://192.168.10.20:3128
  ```

----

## Squid

This forwarder will connect:

* https over a HTTP-connect 'tunnel'
* plain http without such a 'tunnel'

So you need to make sure it is allowed by the proxy:

```text
acl CONNECT method CONNECT
acl ssl_ports port 443
acl step1 at_step SslBump1

http_access deny CONNECT !ssl_ports
http_access allow CONNECT step1
# NOTE: without 'step1' one would be able to create a CONNECT 'tunnel' through the proxy
```

----

## Redirect

### NFTables

Full example when using 'TProxy' mode: [NFTables - TProxy](https://gist.github.com/superstes/6b7ed764482e4a8a75334f269493ac2e)

```bash
# whole input/forward traffic
nft 'add chain nat prerouting { type nat hook prerouting priority -100; }'
nft 'add rule nat prerouting tcp dport { 80, 443 } dnat to 127.0.0.1:3128'

# whole output traffic - excluding the traffic for the proxy-forwarder itself (anti-loop)
nft 'add chain nat output { type nat hook output priority -100; }'
nft 'add rule nat output tcp dport { 80, 443 } meta skuid != 1100 dnat to 127.0.0.1:3128'

# only output-traffic for one user to specific target (nice for testing purposes)
nft 'add rule nat output meta l4proto tcp ip daddr 135.181.170.219 meta skuid 1000 dnat to 127.0.0.1:3128'
```

### IPTables

Full example when using 'TProxy' mode: [IPTables - TProxy](https://gist.github.com/superstes/c4fefbf403f61812abf89165d7bc4000)

```bash
# whole input/forward traffic
sudo iptables -t nat -I PREROUTING -p tcp --dport 80 -j DNAT --to-destination 127.0.0.1:3128
sudo iptables -t nat -I PREROUTING -p tcp --dport 443 -j DNAT --to-destination 127.0.0.1:3128

# whole output traffic - excluding the traffic for the proxy-forwarder itself (anti-loop)
sudo iptables -t nat -I OUTPUT -m owner ! --uid-owner 1100 -p tcp --dport 80 -j DNAT --to-destination 127.0.0.1:3128
sudo iptables -t nat -I OUTPUT -m owner ! --uid-owner 1100 -p tcp --dport 443 -j DNAT --to-destination 127.0.0.1:3128

# only output-traffic for one user to specific target (nice for testing purposes)
sudo iptables -t nat -I OUTPUT -m owner --uid-owner 1000 -d 135.181.170.219 -p tcp -j DNAT --to-destination 127.0.0.1:3128
```
