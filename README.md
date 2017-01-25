# minecraft-rcon

A Golang Minecraft RCON client, for automating RCON commands in non-interactive environments. This
tool was brough about because of a need to automate certain actions without the use of the 
interactive console, the initial specific use-case was for managing a Minecraft server running in a
Docker container.

## Installing

Ensure you have Go installed and set up on your machine, then run the following command:

```
$ go install github.com/SeerUK/minecraft-rcon/... 
```

## Usage

Usage is very straightforward, and options are defined with the default values. Here is an example:

```
$ minecraft-rcon -host 10.0.1.2 -port 35575 -pass correct-horse-battery-staple \
    say Hello, I am using RCON to speak to you right now
```

Any arguments are sent as the command. If a response is given it will be printed out, otherwise 
there should be no output.

## Todo

* Docker container, sent to Docker Hub.
* Tests of some kind
* Versioning

## License

MIT
