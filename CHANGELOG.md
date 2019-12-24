- v? (?):
    + Set default rebuild --workers value to number of cores.
    + Show 'dry run' message when using rebuild --dru-run flag.
    + Implicitly inject 'all' team if not explicitly configured.
    + Add 'identity' option & --identity flag.
    + Add --key flag to 'rebuild' command.
    + Fix rubbish & missing recipients calculation.
    + ...

- v0.3.3 (2019-12-24):
    + Assorted configuration validation fixes.

- v0.3.2 (2019-12-24):
    + Add support for private keys.
    + Validate configuration using JSON schema.
    + Add initial support for Windows.

- v0.3.1 (2019-12-05):
    + Add 'init' command.
    + Add execution hooks.

- v0.3.0 (2019-12-04):
    + Add support for templates (i.e. schemas + skeletons).

- v0.2.3 (2019-12-03):
    + Fix edition when creating new secrets without tags

- v0.2.2 (2019-12-03):
    + Replace permissions regexps by queries.
    + Replace --grep by --query.
    + Add --long to 'ls' command.
    + Remove 'about' command.

- v0.2.1 (2019-12-01):
    + Add support for tags.

- v0.2.0 (2019-12-01):
    + Secrets stored as gzipped files including custom headers.

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
