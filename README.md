# reverseoperator

[![Build Status](https://travis-ci.org/fardog/reverseoperator.svg?branch=master)](https://travis-ci.org/fardog/reverseoperator)
[![](https://godoc.org/github.com/fardog/reverseoperator?status.svg)](https://godoc.org/github.com/fardog/reverseoperator)

A DNS-over-HTTPS server with a Google [DNS-over-HTTPS][dnsoverhttps] compatible
API. Allows you to run your own Google DNS-over-HTTPS compatible server.

This service pairs well with [secure-operator][], which can act as a
DNS-protocol bridge for your local network.

**This service is *alpha quality*.** For now, installing from source is the
only option; once it is of release quality, releases will be provided.

## Installation

Install using `go get`:

```
go get -u github.com/fardog/reverseoperator/cmd/reverse-operator
```

Then either run the built package:

```
reverse-operator
```

This will start an HTTP server listening at `:80`. For usage information, run
`reverse-operator --help`.

**Note:** Running a service on port `80` requires administrative privileges on
most systems. For local development, you may specify a different port using the
`--listen` flag.

## Version Compatibility

This package follows [semver][] for its tagged releases. The `master` branch is
always considered stable, but may break API compatibility. If you require API
stability, either use the tagged releases or mirror on gopkg.in:

```
go get -u gopkg.in/fardog/reverseoperator.v0
```

## Caveats

* No DNS lookup caching is implemented, and likely never will; every request
  will cause a lookup against the configured upstream DNS servers. If you need
  caching, it's up to you to configure a caching DNS server (such as
  [dnsmasq][]) which `reverse-operator` will request against.

## License

```
   Copyright 2017 Nathan Wittstock

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0
```

The Google DNS-over-HTTPS API is licensed under the
[Creative Commons Attribution 3.0 License][cc-by-3.0] license.

[dnsoverhttps]: https://developers.google.com/speed/public-dns/docs/dns-over-https
[cc-by-3.0]: http://creativecommons.org/licenses/by/3.0/
[secure-operator]: https://github.com/fardog/secureoperator
[dnsmasq]: http://www.thekelleys.org.uk/dnsmasq/doc.html
[semver]: https://semver.org/
