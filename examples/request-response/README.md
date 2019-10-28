# request-response
request-response is an example go-parodus client that manages an interanal map of Information.
While the code itself is boring the end2end with XMiDT is awesome
# Running
## prereqs
- Have an XMiDT [cluster up](https://xmidt.io/docs/operating/getting_started/) and running
- Start parodus running on `tcp://127.0.0.1:6666` and connected to the XMiDT cluster with the mac address
of de:ad:be:ef:ca:fe
- run the example `go run .`

## testing
```bash
# set a value in the local map
curl -X POST \
 https://scytale.example.com/api/v2/device \
 -H 'Accept: application/json' \
 -H 'Content-Type: application/json' \
 -H 'X-Webpa-Device-Name: mac:deadbeefcafe/config' \
 -H 'X-Xmidt-Message-Type: 3' \
 -H 'X-Xmidt-Source: dns:foo.bar' \
 -H 'X-Xmidt-Transaction-Uuid: 1234' \
 -d '{
   "command": "SET",
   "parameters": [
       {
           "name": "message",
           "value": "Hello World!",
           "dataType": 0
       }
   ]
}'

#  wrp.message decodedpayload:
#  {"statusCode":200,"message":"Success"}


# get value out of the local map
curl -X POST \
 https://scytale.example.com/api/v2/device \
 -H 'Accept: application/json' \
 -H 'Content-Type: application/json' \
 -H 'X-Webpa-Device-Name: mac:deadbeefcafe/config' \
 -H 'X-Xmidt-Message-Type: 3' \
 -H 'X-Xmidt-Source: dns:foo.bar' \
 -H 'X-Xmidt-Transaction-Uuid: 1234' \
 -d '{
   "command": "GET",
   "names": [
     "message"
   ]
}'

#  wrp.message decodedpayload:
#  {"statusCode":200,"parameters":[{"name":"message","value":"Hello World!","message":"Success"}],"message":"Success"}
```

