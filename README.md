**PGP Tomb is a minimalistic multi-platform command line secrets manager built on top of PGP**. It was created just for fun, mainly as an excuse to play with Go for the first time. Nevertheless, it's fully functional and actively used. Highlights:

- Secrets (i.e. passwords, bank accounts, software licenses, PDF documents, etc.) are stored in the file system as **binary PGP files encrypted & gzipped using one or more public keys** (i.e. recipients). You can create as many tombs (i.e. collections of secrets) as needed.

- Secrets can be **tagged using an unlimited number of (name, value) pairs** stored as unencrypted meta-information of the gzipped file.

- A **flexible permissions model is provided in order to allow sharing secrets in multi-user environments**. Public PGP keys can be organized in teams and access to each secret or collection of secrets can be easily restricted to one or more teams and / or individual users.

- A **flexible JSON Schema based templating model is provided in order to enforce formatting of JSON / YAML secrets** and provide initial skeleton for new secrets. Each secret or group of secrets can optionally and easily be linked to a template in order to enforce your formatting requirements when creating or editing them.

- **Configurable pre & post execution hooks are available in order to ease integration with external tools like Git** (useful to easily share tombs across your organization).

- Encryption of secrets using public ASCII armored PGP keys is directly and efficiently handled by PGP Tomb using the official OpenPGP Go library. However, decryption is built on top of your local GPG infrastructure in order to **seamlessly integrate with your local GPG Agent and avoid messing with your private PGP keys**.

SETUP
=====

1. Download the PGP Tomb executable matching your platform from the [releases page](https://github.com/carlosabalde/releases). Alternatively, for Go >= 1.13, you can run the following command. Finally, you can also use a Docker container (see next section) to build your own executable (i.e. `make build`).
   ```
   $ go get -u github.com/carlosabalde/pgp-tomb/cmd/pgp-tomb/
   ```

2. Somewhere in your file system (e.g. `~/pgp-tomb/`) create the following files & folders. You can do it manually, or simply run `pgp-tomb init ~/pgp-tomb/` to create a brand new tomb:
   1. The PGP Tomb [configuration file](config/pgp-tomb.yaml).
   2. The folder containing the PGP public keys (`.pub` extension and ASCII armor are required) of users in your organization (i.e. no need to import these keys in your local GPG keyring).
   3. The folder containing the [templates](files/templates/) (`.schema` and `.skeleton` extensions are required).
   4. The folder containing the [hooks](files/hooks/) (`.hook` extension and execution permissions are required).
   5. The folder that will store encrypted secrets (`.secret` files will populate this folder once you start using the manager).
   ```
   |-- pgp-tomb.yaml
   |-- hooks/
   |   |-- pre.hook
   |   `-- post.hook
   |-- keys/
   |   |-- alice.pub
   |   |-- bob.pub
   |   `-- chuck.pub
   |-- templates/
   |   |-- login.skeleton
   |   `-- login.schema
   `-- secrets/
   ```

3. Except for permissions and templates, the PGP Tomb configuration is simple and self-explanatory:
   - Permissions for a particular secret are computed matching it (i.e. URI, tags, etc.) against each rule in the configuration. When a match is found, the list of recipients is updated adding (`+` prefix) or removing (`-` prefix) team members / individual users, and then the rule evaluation continues. Obviously order is relevant both for rules as well as for expressions associated to each rule.
   - Users in the list of keepers will always be part of the list of recipients (and at least one keeper is required in a valid configuration).
   - Templates (i.e. JSON Schema and/or JSON / YAML skeletons) are linked to secrets using a similar strategy, however, unlike permissions, evaluation of rules stops once a match is found.
   ```
   root: /home/carlos/pgp-tomb

   keepers:
     - alice

   teams:
     team-1:
       - alice
       - bob
       - chuck
     team-2:
       - chuck

   permissions:
     - uri ~ '^foo/':
         - +team-1
     - uri ~ '^foo/bar/' && tags.type != 'acme':
         - -bob
         - +team-2

   templates:
     - uri ~ '\.login$': login
   ```

4. Optionally you can configure Bash or Zsh completions. For example, for Bash adjust your `.bashrc` as follows.
   ```
   ...
   source /etc/bash_completion
   source <(pgp-tomb bash)
   export PGP_TOMB_ROOT=/home/carlos/pgp-tomb
   #alias pgp-tomb="pgp-tomb --root $PGP_TOMB_ROOT"
   ```

   Same thing in MacOS requires some extra steps: (1) install a modern Bash and the completion extension (e.g. `port install bash bash-completion`); (2) add `/opt/local/bin/bash` to the list of allowed shells in `/etc/shells`; and (3) change your default shell (i.e. `chsh -s /opt/local/bin/bash`).
   ```
   ...
   source /opt/local/etc/profile.d/bash_completion.sh
   source <(pgp-tomb bash)
   export PGP_TOMB_ROOT=/home/carlos/pgp-tomb
   #alias pgp-tomb="pgp-tomb --root $PGP_TOMB_ROOT"
   ```

5. Assuming a local GPG infrastructure properly configured, now you're ready to start creating and sharing secrets across your organization.
   ```
   # Show command line options.
   $ pgp-tomb --help

   # Create a new secret using an editor (you can choose your preferred editor
   # using the EDITOR environment variable).
   $ pgp-tomb edit foo/answers.md

   # Copy contents of a secret to the system clipboard (depends on 'xsel'
   # or 'xclip' in Linux systems).
   $ pgp-tomb get foo/answers.md --copy

   # List URIs of secrets theoretically readable by 'chuck' according with the
   # permissions defined in the current configuration.
   $ pgp-tomb list --long --key chuck

   # Check all secrets and re-encrypt them if current recipients don't match
   # the list of expected recipients according with the current configuration.
   $ pgp-tomb rebuild
   ```

DEVELOPMENT
===========

Running `make docker` you can build & connect to a handy Docker container useful to ease development and packaging phases. This is what you need to know:

- Once connected to the container you can build the project and execute it as follows.
  ```
  # cd /mnt/

  # make build-dev

  # ./build/linux-amd64/pgp-tomb -v -c /mnt/config/pgp-tomb.yaml get 'foo/bar/lorem ipsum.txt'
  ```

- A minimalistic GPG environment is configured including a keyring with a single private key (imported from `/mnt/files/keys/alice.pri`). Additional public & private test key pairs can be found and manually imported from `/mnt/files/keys/` (all private keys encrypted using the same password: `s3cr3t`). Beware that folder is expected to contain just public keys in a real scenario.
  ```
  # gpg --list-secret-keys
  /root/.gnupg/pubring.kbx
  ------------------------
  sec   rsa1024 2019-11-23 [SCEA]
        FA47DEB05289E92E4504962BE94F302B75D78168
  uid           [ unknown] alice <alice@example.com>
  ssb   rsa1024 2019-11-23 [SEA] [expires: 2027-11-21]
  ssb   rsa1024 2019-11-23 [SEA] [expires: 2027-11-21]
  ```

- Some encrypted test files and templates can be found in `/mnt/files/secrets/` and `/mnt/files/templates/` respectively.
  ```
  # zcat /mnt/files/secrets/foo/bar/lorem\ ipsum.txt.secret | gpg --use-agent -d
  ```

- PGP Tomb only support public keys using an ASCII armor. You can adapt existing keys using the following command:
  ```
  # cat 'key.pub' | gpg --enarmor | sed 's|ARMORED FILE|PUBLIC KEY BLOCK|'
  ```

COPYRIGHT
=========

See LICENSE.md for details.

Parsing and evaluation of queries is strongly based on [Binary Expression Tree](https://github.com/alexkappa/exp) project by Alex Kalyvitis.

Copyright (c) 2019 Carlos Abalde <carlos.abalde@gmail.com>
