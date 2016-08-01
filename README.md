# fs-registrator

*FreeSWITCH Sofia-SIP Registry Bridge (Sync to Key/Value Store)*

# Summary

When run and pointed at a FreeSWITCH instance, will take note of all Sofia-SIP registrations, and propagate them to a (replicated) K/V store for lookups.

Useful for discovering which SIP registrations reside on which server/s.

We use ESL events + a semi-regular sync for reconciliation (to gracefully handle restarts and/or missed events).

# Supported K/V Stores

Currently the focus is on [etcd](https://github.com/coreos/etcd), with the intention to support others in future. [Consul](https://github.com/hashicorp/consul) and [redis](https://github.com/antirez/redis) would be the most likely next targets (both support prefix-based lookups).

The architecture is not super pluggable right now, will be abstracted as required.

# Configuration

Configuration is performed via CLI arguments, and self documenting using `--help`:

```

```

# Building

TODO
