# nagios-grafana-parser
A program to fetch alerts from Nagios via API and display it on Grafana
It also provides a (password-protected via LDAP or static conf) Vue.js GUI for managing the fields temporarly:

![](https://i.vgy.me/z1c3Nz.png)
![](https://i.vgy.me/LPngFx.gif)

The configuration is updated automatically on the fly, but it needs to be written to the config file to be made permanent.

The program is able to collect alerts from multiple nagios instances.

How grafana could look (need to fake some issues):
![](https://i.vgy.me/VGcnJz.png)
