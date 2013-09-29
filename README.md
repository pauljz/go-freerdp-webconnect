go-freerdp-webconnect
=====================

A Go port of [FreeRDP/FreeRDP-WebConnect](https://github.com/FreeRDP/FreeRDP-WebConnect) using `cgo` bindings to `libfreerdp` and `go.net/websocket`.

I wrote this port as an experiment to compare with alternatives like [Guacamole](http://guac-dev.org/), after seeing that FreeRDP-WebConnect itself was unmaintained.

The project we were investigating this for only needed an RDP viewer not a fully interactive RDP client, so I've only implemented the display functionality, not the keyboard, touch, and pointer functionality.

This is unmtained and unused but I thought I'd throw it out there for posterity. In the end we decided to use Guacamole.
