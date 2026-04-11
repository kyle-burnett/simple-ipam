# simple-ipam

A small CLI for managing an IP address plan as a hierarchical YAML file. IPv4 only.
Subnets nest under their smallest enclosing parent, each with an optional description and tags.

## Install

```sh
go install github.com/kyle-burnett/simple-ipam@latest
```

Or build from source:

```sh
go build -o simple-ipam .
```

## Commands

| Command | Purpose |
|---|---|
| `init` | Create an empty IPAM file |
| `add` | Add a specific subnet |
| `add-next-available` | Allocate the lowest-addressed free subnet of a given prefix length under a parent |
| `delete` | Delete a subnet (optionally recursive) |

See [`docs/`](docs/) for more details.

## Example

```sh
simple-ipam init -f ipam.yaml -d "corp net"
simple-ipam add -f ipam.yaml -s 10.0.0.0/16 -d "us-east"
simple-ipam add -f ipam.yaml -s 10.0.0.0/24 -d "vpc-a"
simple-ipam add-next-available -f ipam.yaml -p 10.0.0.0/24 -l 26 -d "subnet-a1"
```

The resulting `ipam.yaml`:

```yaml
description: corp net
subnets:
    10.0.0.0/16:
        description: us-east
        tags: []
        subnets:
            10.0.0.0/24:
                description: vpc-a
                tags: []
                subnets:
                    10.0.0.0/26:
                        description: subnet-a1
                        tags: []
                        subnets: {}
```

`add-next-available` picks the lowest free block, reusing holes before appending, and nests the new entry at the deepest existing ancestor.
