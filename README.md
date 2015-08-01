#Patchy

An online jukebox by [Luminarys](https://github.com/Luminarys), [minus](https://github.com/minus7), and [SirCmpwn](https://github.com/SirCmpwn).

![](https://fuwa.se/pip1tc.png)

#Requirements
Patchy requires Go, mpd, and scss. The code itself require the gompd and web packages, so you'll want to run `go get github.com/fhs/gompd/mpd` and `go get github.com/hoisie/web` to grab the necessary libraries.

#Setup
* Setup mpd on your machine so that it runs and is pointing to a music directory.
* Modify patchy.conf to set the default port and mpd music directory locations. Note that these can still be overridden with flag.
* Run `make` to compile all assets and generate the binary.
* Ensure that Nginx or whatever webserver you use is properly configured to handle websockets. An example configuration file for Nginx has been provided in conf which you may examine or use.

#Running
Run `./patchy` to start the server with the default options in patchy.conf. 

You may want to manually specify flags, run `./patchy -h` to see them.

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
