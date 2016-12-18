# cybot
A Twitch bot coded in GoLang

## Prerequisites:

###Docker 
If you're unfamiliar with docker, [go here](https://docs.docker.com/engine/installation/)

### [BurntSushi's Go Toml library](https://github.com/BurntSushi/toml/) (if you want to compile yourself)


I use a docker container with a prebuilt environment and lots of goodies to work in go.
To copy my dev repo, simply cd into the godev directory and run `./build.sh`, then `./run.sh`.
You'll create the container and be placed into it!

# Usage:

Edit the twitch_credentials.toml.template file.  
Fill in the variables. If you don't know how to get your twitch Oauth token, [go here](http://www.twitchapps.com/tmi/)
Rename it to twitch_credentials.toml  

`go build bot.go`  
`./bot`  

