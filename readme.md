This small go program connects to an MQTT server, and checks for temperature data submitted on a specific channel.
If no data is submitted in an hour, a webhook will be invoked (Works with slack).
Also sends a message to the given channel when the status changes for the sensor.

How to run it:
```
go run . \
-host your-mqtt.host.org \
-port 1883 \
-channel living_room/temperature \
-webhook https://hooks.slack.com/services/RANDOM/UUID/FOR_YOUR_BOT
```