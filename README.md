**PGP Tomb is a minimalistic multi-platform command line secrets manager built on top of PGP**. It was created just for fun, mainly as an excuse to play with Go for the first time. Nevertheless, it's fully functional and actively used. Highlights:

- Secrets (i.e. passwords, bank accounts, software licenses, PDF documents, etc.) are stored in the file system as binary PGP files encrypted using one or more public keys (i.e. recipients).

- A simple yet flexible permissions model is provided in order to allow sharing secrets in multi-user environments. Public PGP keys can be organized in teams and access to each secret can be easily restricted to one or more teams and / or individual users. It's up to you how to share the secrets across the organization: a git repository, a shared folder, etc.

- Encryption of secrets using public ASCII armored PGP keys is directly and efficiently handled by PGP Tomb using the official OpenPGP Go library. However, decryption is built on top of your local GPG infrastructure in order to seamlessly integrate with your local GPG Agent and avoid messing with your private PGP keys.

SETUP
=====

1. Download the PGP Tomb executable matching your platform from the [releases page](https://github.com/carlosabalde/pgp-tomb/releases). Alternatively, for Go >= 1.13, you can run the following command. Finally, you can also use a Docker container (see next section) to build your own executable (i.e. `make build`).
   ```
   $ go get -u github.com/carlosabalde/pgp-tomb/cmd/pgp-tomb/
   ```

2. Somewhere in your file system (e.g. `~/pgp-tomb/`) create the following files & folders: (1) the PGP Tomb configuration file; (2) the folder containing the PGP public keys (`.pub` extension and ASCII armor is required) of users in your organization (i.e. no need to import these keys in your local GPG keyring); and (3) the folder that will store encrypted secrets (`.pgp` files will populate this folder once you start using the manager).
   ```
   |-- pgp-tomb.yaml
   |-- keys/
   |   |-- alice.pub
   |   |-- bob.pub
   |   `-- chuck.pub
   `-- secrets/   
   ```

3. Except for permissions, the PGP Tomb configuration is simple and self-explanatory. Permissions for a particular secret are computed matching the URI of the secret (e.g. `foo/answers.md`) against each regexp in the configuration. When a match is found, the list of recipients is updated adding (`+` prefix) or removing (`-` prefix) team members / individual users, and the the pattern matching continues. Obviously order is relevant both for regexps and for expressions associated to each regexp. Users in the list of keepers will always be part of the list of recipients (and a least one keeper is required in a valid configuration).
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
     - ^foo/:
         - +team-1
     - ^foo/bar/:
         - -bob
         - +team-2
   ```

4. Optionally you can configure Bash or Zsh completions. For example, for Bash adjust your `.bashrc` as follows.
   ```
   ...
   source /etc/bash_completion
   source <(pgp-tomb bash)
   export PGP_TOMB_ROOT=/home/carlos/pgp-tomb
   ```

   Same thing in MacOS requires some extra steps: (1) install a modern Bash and the completion extension (e.g. `port install bash bash-completion`); (2) add `/opt/local/bin/bash` to the list of allowed shells in `/etc/shells`; and (3) change your default shell (i.e. `chsh -s /opt/local/bin/bash`).
   ```
   ...
   source /opt/local/etc/profile.d/bash_completion.sh
   source <(pgp-tomb bash)
   export PGP_TOMB_ROOT=/home/carlos/pgp-tomb
   ```

5. Assuming a local GPG infrastructure properly configured, now you're ready to start creating and sharing secrets across your organization.
   ```
   # Show command line options.   
   $ pgp-agent --help

   # Create a new secret using an editor (you can choose your preferred editor
   # using the EDITOR environment variable).
   $ pgp-agent --config ~/pgp-tomb/pgp-tomb.yaml edit foo/answers.md

   # Copy contents of a secret to the system clipboard (depends on 'xsel'
   # or 'xclip' in Linux systems).
   $ pgp-agent --config ~/pgp-tomb/pgp-tomb.yaml get foo/answers.md --copy

   # List URIs of secrets theoretically readable by 'chuck' according with the
   # permissions defined in the current configuration.
   $ pgp-agent --config ~/pgp-tomb/pgp-tomb.yaml list --key chuck

   # Check all secrets and re-encrypt them if current recipients don't match
   # the list of expected recipients according with the current configuration.
   $ pgp-agent --config ~/pgp-tomb/pgp-tomb.yaml rebuild
   ```

DEVELOPMENT
===========

Running `make docker` you can build & connect to a handy Docker container useful to ease development and packaging phases. These is what you need to know:

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

- Some encrypted test files can be found in `/mnt/files/secrets/`.
  ```
  # gpg --use-agent -d < /mnt/files/secrets/foo/bar/lorem\ ipsum.txt.pgp
  ```

- Encrypted files in `/mnt/files/secrets/` match the permissions included in the sample configuration file at `/mnt/config/pgp-tomb.yaml`. They were generated using a temporary keyring as follows.
  ```
  # mkdir /tmp/gnupg/

  # gpg --homedir /tmp/gnupg --import /mnt/files/keys/*.pub

  # gpg --homedir /tmp/gnupg --yes --encrypt --compress-algo 1 \
        --output /mnt/files/secrets/foo/Lenna.png.pgp \
        --recipient alice@example.com \
        --recipient bob@example.com \
        --recipient frank@example.com \
        /mnt/files/plain/Lenna.png

  # gpg --homedir /tmp/gnupg --yes --encrypt --compress-algo 1 \
        --output '/mnt/files/secrets/foo/bar/Mrs Dalloway.txt.pgp' \
        --recipient alice@example.com \
        --recipient bob@example.com \
        --recipient frank@example.com \
        '/mnt/files/plain/Mrs Dalloway.txt'

  # gpg --homedir /tmp/gnupg --yes --encrypt --compress-algo 1 \
        --output '/mnt/files/secrets/foo/bar/lorem ipsum.txt.pgp' \
        --recipient alice@example.com \
        --recipient bob@example.com \
        --recipient chuck@example.com \
        '/mnt/files/plain/lorem ipsum.txt'

  # gpg --homedir /tmp/gnupg --yes --encrypt --compress-algo 1 \
        --output /mnt/files/secrets/quz/answers.md.pgp \
        --recipient alice@example.com \
        --recipient bob@example.com \
        /mnt/files/plain/answers.md

  # rm -rf /tmp/gnupg/
  ```

- PGP Tomb only support public keys using an ASCII armor. You can adapt existing keys using the following command:
  ```
  # cat 'key.pub' | gpg --enarmor | sed 's|ARMORED FILE|PUBLIC KEY BLOCK|'
  ```

COPYRIGHT
=========

See LICENSE.md for details.

Copyright (c) 2019 Carlos Abalde <carlos.abalde@gmail.com>
