#Patchy

An online jukebox by [Luminarys](https://github.com/Luminarys), [minus](https://github.com/minus7), and [SirCmpwn](https://github.com/SirCmpwn).

![](https://fuwa.se/pip1tc.png)

#Requirements
Patchy requires Go and mpd. You'll want to run `go get github.com/fhs/gompd/mpd` to grab the library that's used to interface mpd.

#Setup
First setup mpd properly on your machine. It just has to point to be told where your music directory is and given some interface to play into. This can be a dummy interface for a headless server or something like an http stream. 
Then modify src/main.go and change the musicDir variable to point to your mpd music directory. You can then run `make` to compile all assets and generate the binary. 

#Running
Run `./patchy`, and you should be good to go. Please note that it currently must be run from within the git repo or else it will not work.

#Features:
* Music library searching
* Precise stream synchronization via websockets
* Client queue requests
* Client music uploads
* Use a configuration file for stuff

#Things to Do:
* Fix font errors for Windows to Linux ULs
* Deuglify stuff
* Equalizer(?)
