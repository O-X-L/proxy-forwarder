# Proxy Forwarder

This tool is specifically designed to solve a problem when using proxy server:

* There is no clean way of forwarding traffic to a remote proxy server that is outside your layer 2 network
* Some proxy servers (_like Squid_) do not support redirecting the traffic using DNAT


<a href="https://wiki.superstes.eu/en/latest/1/network/squid.html#transparent-proxy">
<img src="https://github.com/superstes/proxy-forwarder/blob/latest/docs/squid_remote.png" alt="Remote Proxy Server" width="600"/>
</a>

----

## How does it work?

This tool is based on top of [go-gost](https://gost.run/en/tutorials/redirect/) but I stripped all unecessary features/functions/dependencies from it.

See also: [gost documentation](https://wiki.superstes.eu/en/latest/1/network/gost.html)

### Usage

```bash
  -P 'Listen port' (required)
  -F 'Proxy server to forward the traffic to' (required, Example: 'http://192.168.0.1:3128')
  -T 'Run in TProxy mode' (default: false)
  -M 'Mark to set for TProxy traffic' (default: 100)
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
  bash scripts/build.sh
  ```

* Copy the new binary to the target host(s)
* Run

  ```bash
  proxy_forwarder -P 4138 -F http://127.0.0.1:3128
  ```
