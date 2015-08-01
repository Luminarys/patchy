#Patchy

An online jukebox by [Luminarys](https://github.com/Luminarys), [minus](https://github.com/minus7), and [SirCmpwn](https://github.com/SirCmpwn).

![](https://fuwa.se/pip1tc.png)

#Requirements
Patchy requires Go, mpd, and scss. The code itself require the gompd and web packages, so you'll want to run `go get github.com/fhs/gompd/mpd` and `go get github.com/hoisie/web` to grab the necessary libraries.

#Setup
First setup mpd properly on your machine. It just has to point to be told where your music directory is and given some interface to play into. This can be a dummy interface for a headless server or something like an http stream. 
You can then modify patchy.conf to set the default port and mpd music directory locations. Note that these can still be overridden with flag. You can then run `make` to compile all assets and generate the binary.
You must also ensure that Nginx is properly configured to handle websockets. An example configuration file has been provided in conf which you may examine or use.

#Running
Run `./patchy`, and you should be good to go. You can also examine and change the flags, which are set to defaults by the values in patchy.conf. Please note that patchy currently must be run from within the git repo or else it will not work.

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
