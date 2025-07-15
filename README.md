### compile go script ###
go build mqtt.go


### create service ###
sudo vim /etc/systemd/system/vectorMQTT.service


### vectorMQTT.service have to contain the following ####
[Unit]
Description=Vector MQTT script
After=network.target
Requires=network.target

[Service]
ExecStartPre=/bin/sleep 30
ExecStart=/home/pi/VectorApps/plugins/mqtt SERIAL_NO1 SERIAL_NO2 SERIAL_NO3
Restart=always
RestartSec=30
User=pi

[Install]
WantedBy=multi-user.target


### start service ####
sudo systemctl daemon-reload
sudo systemctl enable vectorMQTT.service
sudo systemctl start vectorMQTT.service
