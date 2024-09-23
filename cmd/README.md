Package: main

Imports:
- flag

External data, input sources:
- Command-line flags

## Parsing Flags

This function parses command-line flags and returns a configuration struct. It initializes a new config struct and then uses the flag package to define and parse the following flags:

- `rendezvous`: A string that identifies a group of nodes. This flag is used to connect with friends.
- `wrapped_host`: The bootstrap node's wrapped host listen address.
- `pid`: Sets a protocol id for stream headers.
- `port`: The node's listen port.

The function parses the flags using `flag.Parse()` and returns the populated config struct.

