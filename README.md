# Proxy Forwarder

This tool is specifically designed to solve a problem when using proxy servers:

* Setting the environment-variables 'HTTP_PROXY', 'HTTPS_PROXY', 'http_proxy' and 'https_proxy' for all applications and HTTP-clients may be problematic/too inconsistent
* There is no clean way of forwarding all system traffic to a remote proxy server that is outside your layer 2 network
* Some proxy servers (_like Squid_) do not support redirecting the traffic using DNAT


<a href="https://wiki.superstes.eu/en/latest/1/network/squid.html#transparent-proxy">
<img src="https://github.com/superstes/proxy-forwarder/blob/latest/docs/squid_remote.png" alt="Remote Proxy Server" width="600"/>
</a>

For more information about Squid see: [Superstes Wiki - Squid](https://wiki.superstes.eu/en/latest/1/network/squid.html)

----

## How does it work?

This tool is based on [go-gost](https://gost.run/en/tutorials/redirect/) but was stripped of all features/dependencies that are unnecessary to perform this task.

See also: [gost documentation](https://wiki.superstes.eu/en/latest/1/network/gost.html)

### Usage

```bash
  -P 'Listen port' (required)
  -F 'Proxy server to forward the traffic to' (required, Example: 'http://192.168.0.1:3128')
  -T 'Run in TProxy mode' (default: false)
  -M 'Mark to set for TProxy traffic' (default: None)
  -V 'Show version'
  -D 'Enable debug mode'
  -metrics 'Set a metrics service address (prometheus)' (Example: '127.0.0.1:9000', Docs: 'https://gost.run/en/tutorials/metrics/')
  -no-log-time 'Do not add timestamp to logs'  # use when systemd service
```

### It does

* Bind to localhost (_127.0.0.1 & ::1_) for tcp & udp
* Allow you to redirect traffic to the forwarder using:

  * Destination NAT (_default_)
  * or [TProxy Mode](https://docs.kernel.org/networking/tproxy.html) (see also: [extended documentation](https://wiki.superstes.eu/en/latest/1/network/nftables.html#tproxy))

* Forward the traffic to the server defined using the `-F` flag


These are the main two files that cover the logic:

* [redirect-tcp handler](https://github.com/superstes/proxy-forwarder/blob/latest/gost/x/handler/redirect/tcp/handler.go) (*HTTP/HTTPS split*)
* [http connector](https://github.com/superstes/proxy-forwarder/blob/latest/gost/x/connector/http/connector.go) (*CONNECT Tunnel*)

----


## Build

* [Install go](https://go.dev/doc/install)
* Download this repository
* Build the binary

  ```bash
  env GOOS=linux GOARCH=amd64 bash scripts/build.sh

  # use 'CGO_ENABLED' if you see this error on the target system:
  > error "/lib/x86_64-linux-gnu/libc.so.6: version `GLIBC_2.32'"

  env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 bash scripts/build.sh
  ```

* Copy the new binary to the target host(s)
* Run

  ```bash
  proxy_forwarder -P 4138 -F http://192.168.10.20:3128
  ```

----

## Examples

```bash
> curl https://superstes.eu
# proxy-forwarder
2023-08-29 20:49:10 | INFO | handler | 192.168.11.104:36386 <=> superstes.eu:443/tcp | connection established
# squid
NONE_NONE/200 0 CONNECT superstes.eu:443 - HIER_NONE/- -
TCP_TUNNEL/200 6178 CONNECT superstes.eu:443 - HIER_DIRECT/superstes.eu -

> curl http://superstes.eu
# proxy-forwarder
2023-08-29 20:49:07 | INFO | handler | 192.168.11.104:50808 <=> superstes.eu:80/tcp | connection established
# squid
TCP_REFRESH_MODIFIED/301 477 GET http://superstes.eu/ - HIER_DIRECT/superstes.eu text/html
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

----

## Permissions

### Destination NAT

If you use DNAT to redirect the traffic - the service user running the proxy-forwarder has no need for additional privileges!

A default linux user with `/usr/sbin/nologin` as shell is enough.

### TPROXY

If you want to use TPROXY to redirect the traffic - the service user needs the privilege to set `cap_net_raw` on its sockets.

The [CAP_NET_RAW](https://man7.org/linux/man-pages/man7/capabilities.7.html) may be needed for this:

> bind to any address for transparent proxying

You can add it like this:

```bash
setcap cap_net_raw=+ep /usr/local/bin/proxy-forwarder

# make sure only wanted users can execute the binary!
useradd proxy_forwarder --shell /usr/sbin/nologin
chown root:proxy_forwarder /usr/local/bin/proxy-forwarder
chmod 750 /usr/local/bin/proxy-forwarder
```

----

## Service

Here's an example systemd service to run the forwarder:

```text
# /etc/systemd/system/proxy-forwarder.service

[Unit]
Description=Proxy forwarder
Documentation=https://github.com/superstes/proxy-forwarder
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/proxy_forwarder -P 4128 -F http://192.168.1.20:3128 -no-log-time
User=proxy_forwarder
Group=proxy_forwarder
Restart=on-failure
RestartSec=5s

StandardOutput=journal
StandardError=journal
SyslogIdentifier=proxy_forwarder

[Install]
WantedBy=multi-user.target
```
