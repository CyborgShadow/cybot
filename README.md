# cybot
A Twitch bot coded in GoLang

# Usage:

You can either:
## [Download and use our Prebuilt binaries](https://github.com/CyborgShadow/cybot/releases)

## Build it yourself

### Option 1: Use our Docker Container

If you're unfamiliar with docker, [go here](https://docs.docker.com/engine/installation/)

I use a docker container with a prebuilt environment and lots of goodies to work in go.  
To copy my dev repo, simply cd into the godev directory and run `./build.sh`, then `./run.sh`.  
It'll create the container and be placed into it!

Edit the twitch_credentials.toml.template file.  
Fill in the variables. If you don't know how to get your twitch Oauth token, [go here!](http://www.twitchapps.com/tmi/)  
Rename it to twitch_credentials.toml  

Run: `./build_linux.sh`  


Then run it with:  
`./cybot`    

### Option 2: Build your own dev environment 

### You'll need
[BurntSushi's Go Toml library](https://github.com/BurntSushi/toml/)  
[inconshreveable's log15 go logger](https://github.com/inconshreveable/log15)  
[sorcix irc controller](https://github.com/sorcix/irc)  
[mudler's sendfd](https://github.com/mudler/sendfd)  

Depending upon your OS, run one of:  
Linux: `./build_linux.sh`  
Mac: `./build_darwin.sh`  
Windows: `./build_windows.bat`  

Then run it with:  
`./cybot` (`cybot.exe` for windows)  
