# Spartakus User Documentation

If you prefer a hands-on experience over reading a lot of text, check out the [Spartakus walk-through](walk-through.md), otherwise read on.

## Terminology

In Spartakus we differentiate between two roles (both supported by the same program):

- A **volunteer** periodically generates reports using the Kubernetes API and publishes it to the collector.
- A **collector** receives reports from volunteers and stores them in a reporting back-end (default: Google BigQuery).

Each cluster you're reporting on is uniquely identified by a user-provided cluster identifier.

## Commands

## Extensions

Reports can be extended to include additional, custom information called **extensions**. Extensions are key-value pairs with the following requirements:

- Valid keys have two segments: an optional prefix and a name, separated by a slash (`/`). The name segment is required and the prefix is optional. If specified, the prefix must be a DNS sub-domain: a series of DNS labels separated by dots (`.`), e.g. "example.com".
- The values are not restricted in terms of structure or content.

In order to submit extensions, run the volunteer with the `--extensions` flag.
This flag should be set to the path of a file of arbitrary JSON key-value pairs
you would like to report. For example:

```bash
$ spartakus volunteer --cluster-id=$(uuidgen) \
                      --extensions=/path/to/extensions.json
```

With `extensions.json` for example like:

```json
{
  "example.com/hello": "world",
  "example.com/foo": "bar"
}
```

Using above `extensions.json` as configuration, the volunteer will generate a report that looks like something like:

```json
{
    "version": "v1.0.0",
    "timestamp": "867530909031976",
    "clusterID": "2f9c93d3-156c-47aa-8802-578ffca9b50e",
    "masterVersion": "v1.3.5",
    "extensions": [
      {
        "name": "example.com/hello",
        "value": "world"
      },
      {
        "name": "example.com/foo",
        "value": "bar"
      }
    ]
}
```

Note that the `--extensions` flag can optionally be set to the path of a directory. In this case, all files in the provided directory, excluding those with a leading
`.`, will be parsed.

## Security considerations

If you're using Spartakus in a cluster with RBAC enabled, you will have to create a role and a role binding as follows so that Spartakus has the appropriate permissions (essentially allowed to list nodes):

```bash
$ kubectl create role nodelister \
        --verb=get --verb=list \
        --resource=nodes

$ kubectl create rolebinding nodelisterbinding \
        --role=nodelister \
        --serviceaccount=default:default
```

Note that above assumes you're running Spartakus in the default namespace.
