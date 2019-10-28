# events
events is an example go-parodus client that spam the XMiDT cluster with the date to the go-parodus destination
While the code itself is boring the end2end with XMiDT is awesome
# Running
- Have an XMiDT [cluster up](https://xmidt.io/docs/operating/getting_started/) and running
- Have a wrp-listern[https://github.com/xmidt-org/wrp-listener] running and listening for the event `go-parodus`
- Start parodus running on `tcp://127.0.0.1:6666` and connected to the XMiDT cluster with the mac address
of de:ad:be:ef:ca:fe
- run the example `go run .`

The wrp-listner should be printing out consecutive timestamps

