/*
Package joysticks, provides simplified event routing, by channels, from the Linux joystick driver File-like interface.

events can be listened for from any thread, re-routed and simulated.

*/
package joysticks

/*
usage details.

connect to a joystick by index number then use functions to make event servicing channels, one for each button or hat.

event channels provide at least time.

hat channel event provides current position, (x,y) (with only one axis changing per event.)

*/
