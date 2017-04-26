/*
Package joysticks, provides simplified event routing, through channels, from the Linux joystick driver File-like interface.

events can be listened for from any thread, dynamically re-mapped and simulated.

Highlevel Usage

'Capture', one call to setup and start basic 'Event' channeling on the first available device.

Midlevel

'Connect(index)' to a HID.

Use methods to add (or alter) 'Event' channels.

Start running by calling 'ParcelOutEvents()'.

(unlike highlevel, event index to channel mappings can be changed dynamically.)

Lowlevel

'Connect' to a HID by index number.

handle all events directly appearing on the returned HID's OSEvents channel.

Interface

'Event' interface, provides a time.Duration through a call to the Moment() method, returning whatever the underlying Linux driver provides as the events timestamp, as a time.Duration.

returned 'Event's need asserting to their underlying type ( '***Event' ) to access data other than moment.

*/
package joysticks
