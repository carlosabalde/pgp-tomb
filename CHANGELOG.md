- v0.1.2 (?):
    + Secrets (.pgp -> .secret) nos stored as gzipped files including custom headers.
    + ...

- v0.1.1 (2019-11-29):
    + Fix encryption of binary contents.

- v0.1.0 (2019-11-28):
    + Add --force & --workers flags to 'rebuild' command.

- v0.0.8 (2019-11-28):
    + Always use secret URI completion as fallback in Bash completion.
    + Improved internals and error handling.

- v0.0.7 (2019-11-28):
    + Rename --limit flag to --grep in 'list' & 'rebuild' commands.
    + Add 'folder' positional argument to 'list' & 'rebuild' commands.

- v0.0.6 (2019-11-27):
    + Add Bash & Zsh completions.

- v0.0.5 (2019-11-27):
    + Include git hash in --version output.

- v0.0.4 (2019-11-27):
    + Look for configuration also in --root, when provided.
    + Add some useful command aliases.

- v0.0.3 (2019-11-26):
    + Use current folder instead of binary folder as default root.
    + Enforce umask 0777.

- v0.0.2 (2019-11-26):
    + Cosmetic.

- v0.0.1 (2019-11-26):
    + Initial release.
