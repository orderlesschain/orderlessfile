## Creating Keys and Certificates for nodes and clients

Tutorial for creating the cert.pem and key.pem: [Link](https://blog.dnsimple.com/2017/08/how-to-test-golang-https-services/)
Also details: [Link](https://blog.gopheracademy.com/advent-2019/go-grps-and-tls/)


Commands:

go run  ./certificates/generate_cert.go --rsa-bits 1024 --ca --start-date "Jan 1 00:00:00 1970" --duration=1000000h --host 127.0.0.1,localhost
