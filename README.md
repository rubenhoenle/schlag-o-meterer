# Schlag-O-Meterer

```bash
# running the application
nix run

# get the current counter value
ssh -p 23235 127.0.0.1 "get"

# set the new counter value
ssh -p 23235 127.0.0.1 "set 65"

# increment the counter value
ssh -p 23235 127.0.0.1 "incr 10"
```

## Building and running the container image

```bash
# build the container image
nix build .#containerimage

# load the container image into docker
docker load < result

# start the container
docker compose up -d
```

## Configuration

- `SSH_HOST`: ssh host to use (default `localhost`)
- `SSH_PORT`: ssh port to use (default `23235`)
- `SSH_PUBKEY_FILE`: path to file with ssh public keys (`ed25519` key type, seperated by `\n`) which are allowed to increment / edit the counter (default `./.pubkeys`). If the file isn't found the application will start, but nobody will be able to modify the counters value.

Example for the `SSH_PUBKEY_FILE` content

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAS9nLidDyoFWspHE/IFB7ULMnsLsM+YtGKYreYH7UTa
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEzsOgEiuiTQEUZnMORRmhMHDSAo8VBUl/g55Ec6ZaKM
```
