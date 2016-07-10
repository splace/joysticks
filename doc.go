/*
Package joysticks, provides simplified event routing, through channels, from the Linux joystick driver File-like interface.

events can be listened for from any thread, re-routed and simulated.

usage:

'Capture', a single call to setup and start basic event routing.

or (more flexible)

'Connect' to a joystick by index number, then use methods to add event channels, one for each button or hat, and start running by calling 'ProcessEvents'.

event channels provide at least time. event is an interface with a 'Moment' method which returns a time.Duration.

event 'Moment' returns whatever the underlying Linux driver provides as the events timestamp, in time.Duration.

hat channel event provides current position, (x,y) the event will need casting to the hat event to access these. (with only one axis changing per event.)

*/
package joysticks

