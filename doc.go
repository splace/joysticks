/*
Package joysticks, provides simplified event routing, through channels, from the Linux joystick driver File-like interface.

events can be listened for from any thread, dynamically re-mapped and simulated.

usage:

'Capture', a single call to setup and start basic 'Event' channeling on the first available device.

'Event', an interface, provides a time.Duration through the Moment() method, returning whatever the underlying Linux driver provides as the events timestamp, as a time.Duration.

Events, will need casting to the actual type to access data other than moment. 


or (more flexible)

'Connect(index)' to a HID. 

Use methods to add (or alter) 'Event' channels, 

Start running by calling 'ParcelOutEvents()'.

event index to channel mappings can be changed dynamically.

or (DIY)

'Connect' to a HID by index number

handle all events directly using the returned HID's OSEvent channel.

*/
package joysticks
