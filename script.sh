   
#!/bin/sh
# This script gets the gateway address and ports.

   echo "Please enter the gateway IP:"
   read gatewayIP
   echo "Please enter the Synchronization mode"
   echo "0 - Clock Sync or 1 - Logical Clocks"
   read clock
   echo "Welcome to UMASS $gatewayIP $clock"

osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/gateway/gateway using $clock"'
sleep 5s
#ping $gatewayIP
osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/user/user using \"192.168.0.102\""'
#osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/temperaturesensor/temperaturesensor gatewayIP"'
#osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/smartoutlet/smartoutlet gatewayIP"'
#osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/motionsensor/motionsensor gatewayIP"'
#osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/smartbulb/smartbulb gatewayIP"'
#osascript -e 'tell application "Terminal" to do script "/Users/ameetrivedi/Desktop/Amee/Projects/CS677/Lab2/src/github.com/ppegusii/cs677-smart-homes-IoT/doorsensor/doorsensor gatewayIP"'

# Do not have gnome-terminal provision on mac to verify this
# http://www.linuxjournal.com/article/3065

#gnome-terminal -x sh -c './gateway/gateway; exec bash'
#gnome-terminal -x sh -c 'command2; exec bash'
#gnome-terminal -x sh -c 'command3; exec bash'
#gnome-terminal -x sh -c 'command4; exec bash'
#gnome-terminal -x sh -c 'command5; exec bash'
#gnome-terminal -x sh -c 'command6; cexec bash'