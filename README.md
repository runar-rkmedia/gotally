# GoTally

At the moment, this is another implementation of my older game Tally Board, which I released for iOS in 2014.

## About the game

The game is inspired by 2048, but the player sees more mathematically aligned challenges.

## Getting started 

### Dependencies:

- go (for the api)
- node and npm (for the frontend)
- (optional) [fd](https://github.com/sharkdp/fd) and [entr](https://github.com/eradman/entr) for filewatching
- [buf CLI](https://docs.buf.build/installation) for generating the buf

## Starting with makefile

simply run `make start`. 

This will start the server, and also the frontend-dev-server.

### Watch-mode

run `make watch`

Note that this also runs the server-tests. 

### Frontend-development using the public server.

run `make web_public_api`.

The frontend will now work towards the public api (https://gotally.fly.io).

For normal development, it is adviced to use the local server, as it does not pollute the public database.

### Running tests

Ensure the local server is running, as well as the frontend.

run `make test`


## About this project

This project now serves more as a playground for me testing some new technology.

I am planning to create a package here to perhaps test the  usage of [Buf](https://buf.build/), just for kicks.
