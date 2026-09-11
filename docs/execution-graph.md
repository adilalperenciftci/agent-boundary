# Execution graph

The graph is a deterministic projection of verified events, not a separate source of truth.
Nodes use composite process keys. Edges state only what the event supports:

- `observed_exec_parent`: parent identity was associated with child at exec;
- `file_open_sensitive`: process opened a categorized sensitive path;
- `file_open_output`: process opened an output for writing;
- `file_rename_output`: process renamed a path into the output root;
- `network_connect`: process attempted a connection to a numeric destination;
- `artifact_finalized`: collector associated final artifact digest with an observed process.

No edge is named `caused`. A write-open does not prove bytes were written, and ancestry does
not prove semantic influence. Nodes and edges are sorted before hashing, so replay yields the
same graph digest for the same verified stream.
