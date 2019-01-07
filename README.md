![2019-01-07-205954_1366x768_scrot](https://raw.githubusercontent.com/ddidwyll/comico/master/other/2019-01-07-205954_1366x768_scrot.png)
![2019-01-07-210112_1366x768_scrot](https://raw.githubusercontent.com/ddidwyll/comico/master/other/2019-01-07-210112_1366x768_scrot.png)
![2019-01-07-210152_1366x768_scrot](https://raw.githubusercontent.com/ddidwyll/comico/master/other/2019-01-07-210152_1366x768_scrot.png)
![2019-01-07-210306_1366x768_scrot](https://raw.githubusercontent.com/ddidwyll/comico/master/other/2019-01-07-210306_1366x768_scrot.png)

Install (# - from root, $ - from comico):

    # yum install golang nodejs ImageMagic
    # useradd -d /var/www/comico -U comico
    # passwd comico
    # login comico
    $ mkdir ~/go/src
    $ cd ~/go/src
    $ git clone https://github.com/ddidwyll/comico.git
    $ cd comico
    $ cp other/config.json.example ./config.json
    $ vi ./config.json
    $ go get
    $ go build
    $ cd client
    $ npm i
    $ npm run build
    $ exit
    # cp /var/www/comico/go/src/comico/other/comico.service /etc/systemd/system/
    # systemctl daemon-reload
    # systemctl start comico

If port in config.json == "80" || "443" comico use ACME cert

If no config.json presents check http://localhost:8080

For Ubuntu, Debian, Elementary, Mint use apt. Etc

After install login admin|admin and enjoy :)

Any questions <ddidwyll@gmail.com>
