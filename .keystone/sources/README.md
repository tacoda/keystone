# Sources

In 4.x, this folder held `kind: source` files. Each one named external knowledge for stage 3 of runtime resolution, such as a docs folder. `keystone migrate` moved it here from `.charter/sources/`.

Keystone 5.0 has no `source` kind. It doesn't index or read the files in this folder, and `keystone new source` reports an unknown kind.

To give the agent a callable over an external system, write a tool in `.agents/tools/`. For a reference the agent should read, write a reference in `.agents/references/`. See [`docs/kinds/README.md`](../../docs/kinds/README.md) for every kind.
