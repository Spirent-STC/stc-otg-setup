package gosnappi_examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

// PORT1=10.109.121.181/1/1 PORT2=10.109.123.254/1/1 OTGSERVER=localhost:50051 go test -v -test.run TestGtpv1

func TestGtpv1(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	api := gosnappi.NewApi()

	// Set the transport protocol to HTTP
	api.NewGrpcTransport().
		SetLocation(OTGSERVER).
		SetDialTimeout(3 * time.Minute).
		SetRequestTimeout(10 * time.Minute)

	// Create a new traffic configuration that will be set on OTG
	config := gosnappi.NewConfig()

	// Add a test port to the configuration
	ptx := config.Ports().Add().SetName("port1").SetLocation(PORT1)
	prx := config.Ports().Add().SetName("port2").SetLocation(PORT2)

	dev_tx := config.Devices().Add().SetName(fmt.Sprintf("%s_DEV", ptx.Name()))
	eth_tx := dev_tx.Ethernets().Add().SetName(fmt.Sprintf("%s_ETH", ptx.Name())).SetMac("00:11:22:33:44:66").SetMtu(1000)
	eth_tx.Vlans().Add().SetName("vlan10").SetId(10).SetPriority(2).SetTpid("x9300")
	eth_tx.Connection().SetPortName(ptx.Name())
	// eth_tx.SetPortName(ptx.Name())

	dev_rx := config.Devices().Add().SetName(fmt.Sprintf("%s_DEV", prx.Name()))
	eth_rx := dev_rx.Ethernets().Add().SetName(fmt.Sprintf("%s_ETH", prx.Name())).SetMac("00:11:22:33:44:55").SetMtu(1000)
	eth_rx.Vlans().Add().SetName("vlan20").SetId(20).SetPriority(2).SetTpid("x9300")
	eth_rx.Connection().SetPortName(prx.Name())
	// eth_rx.SetPortName(prx.Name())

	// Configure a flow and set previously created test port as one of endpoints
	flow := config.Flows().Add().SetName("flow1")
	flow.TxRx().Port().SetTxName(ptx.Name()).SetRxNames([]string{prx.Name()})
	// and enable tracking flow metrics
	flow.Metrics().SetEnable(true)

	// Configure number of packets to transmit for previously configured flow
	flow.Duration().Continuous()
	// and fixed byte size of all packets in the flow
	flow.Size().SetFixed(128)
	flow.Rate().SetPercentage(1.0)

	// Configure protocol headers for all packets in the flow
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv4 := pkt.Add().Ipv4()
	//tcp := pkt.Add().Tcp()
	udp := pkt.Add().Udp()
	gtpv1 := pkt.Add().Gtpv1()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv4.Src().SetValue("10.1.1.1")
	ipv4.Dst().SetValue("20.1.1.1")

	udp.SrcPort().SetValue(2152)
	//udp.SrcPort().SetValues([]uint32{100, 1000, 10000, 65535})
	//udp.SrcPort().Increment().SetStart(1).SetStep(64).SetCount(100)
	//udp.SrcPort().Decrement().SetStart(65535).SetStep(128).SetCount(10)
	udp.DstPort().SetValue(2152)
	//udp.DstPort().SetValues([]uint32{100, 1000, 10000, 65535})
	//udp.DstPort().Increment().SetStart(1).SetStep(64).SetCount(100)
	//udp.DstPort().Decrement().SetStart(65535).SetStep(128).SetCount(10)
	//udp.Length().SetValue(65534)
	//udp.Length().SetValues([]uint32{2, 10, 100, 1000, 10000, 65534, 65535})
	//udp.Length().Increment().SetStart(1).SetStep(1).SetCount(1000)
	//udp.Length().Decrement().SetStart(65535).SetStep(1).SetCount(1000)
	//udp.Checksum().SetGenerated("bad")
	//udp.Checksum().SetCustom(65534)
	gtpv1.SquenceNumber().SetValue(100)
	//gtpv1.ExtensionHeaders().Add().ExtensionLength().SetValue(100)
	gtpv1.ExtensionHeaders().Add().ExtensionLength().SetValues([]uint32{10, 100})
	//gtpv1.ExtensionHeaders().Add().ExtensionLength().Increment().SetStart(1).SetStep(1).SetCount(10)
	//gtpv1.ExtensionHeaders().Add().ExtensionLength().Decrement().SetStart(100).SetStep(1).SetCount(10)

	//gtpv1.ExtensionHeaders().Add().NextExtensionHeader().SetValue(255)

	//gtpv1.ExtensionHeaders().Add().NextExtensionHeader().SetValues([]uint32{0, 10, 100, 200, 255})
	//gtpv1.ExtensionHeaders().Add().NextExtensionHeader().Increment().SetStart(0).SetStep(1).SetCount(254)
	//gtpv1.ExtensionHeaders().Add().NextExtensionHeader().Decrement().SetStart(255).SetStep(1).SetCount(254)
	gtpv1.ExtensionHeaders().Add().Contents().SetValue(1234)
	//gtpv1.ExtensionHeaders().Add().Contents().SetValues([]uint64{4660, 273, 546})
	//gtpv1.ExtensionHeaders().Add().Contents().Increment().SetStart(273).SetStep(0001).SetCount(10)
	//gtpv1.ExtensionHeaders().Add().Contents().Decrement().SetStart(273).SetStep(0001).SetCount(10)
	// gtpv.Version().SetValue(1)
	// gtpv.Version().SetValues([]uint32{1})
	// gtpv.Version().Increment().SetStart(1).SetStep(1).SetCount(1)
	gtpv1.Version().Decrement().SetStart(1).SetStep(1).SetCount(1)
	// gtpv.ProtocolType().SetValue(1)
	// gtpv.ProtocolType().SetValues([]uint32{1})
	// gtpv.ProtocolType().Increment().SetStart(1).SetStep(1).SetCount(1)
	// gtpv.ProtocolType().Decrement().SetStart(1).SetStep(1).SetCount(1)
	// gtpv.Reserved().SetValue(0)
	// gtpv.Reserved().SetValues([]uint32{0})
	// gtpv.Reserved().Increment().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.Reserved().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.sFalg().SetValue(0)
	// gtpv.PnFlag().SetValue(0)
	// gtpv.PnFlag().SetValues([]uint32{0})
	// gtpv.PnFlag().Increment().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.PnFlag().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.MessageType().SetValue(0)
	// gtpv.MessageType().SetValues([]uint32{0})
	// gtpv.MessageType().Increment().SetStart(0).SetStep(1).SetCount(1)
	gtpv1.MessageType().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.Teid().SetValue(0)
	// gtpv.Teid().SetValues([]uint32{0})
	// gtpv.Teid().Increment().SetStart(0).SetStep(1).SetCount(1)
	gtpv1.Teid().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.MessageLength().SetValue(0)
	// gtpv.MessageLength().SetValues([]uint32{0})
	// gtpv.MessageLength().Increment().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.MessageLength().Decrement().SetStart(0).SetStep(1).SetCount(1)

	// gtpv.NPduNumber().SetValue(0)
	// gtpv.NPduNumber().SetValues([]uint32{0})
	// gtpv.NPduNumber().Increment().SetStart(0).SetStep(1).SetCount(1)
	gtpv1.NPduNumber().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.NextExtensionHeaderType().SetValue(0)
	// gtpv.NextExtensionHeaderType().SetValues([]uint32{0})
	// gtpv.NextExtensionHeaderType().Increment().SetStart(0).SetStep(1).SetCount(1)
	gtpv1.NextExtensionHeaderType().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.SquenceNumber().SetValue(0)
	// gtpv.SquenceNumber().SetValues([]uint32{0})
	// gtpv.SquenceNumber().Increment().SetStart(0).SetStep(1).SetCount(1)
	gtpv1.SquenceNumber().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.EFlag().SetValue(0)
	// gtpv.EFlag().SetValues([]uint32{0})
	// gtpv.EFlag().Increment().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.EFlag().Decrement().SetStart(0).SetStep(1).SetCount(1)
	// gtpv.SFlag().SetValue(0)
	// gtpv.SFlag().SetValues([]uint32{0})
	// gtpv.SFlag().Increment().SetStart(0).SetStep(1).SetCount(1)

	fmt.Println("Test Gosnappi begin :")
	// // Configure repeating patterns for source and destination UDP ports
	// udp.SrcPort().SetValues([]uint32{5010, 5015, 5020, 5025, 5030})
	// udp.DstPort().Increment().SetStart(6010).SetStep(5).SetCount(5)
	//tcp.SrcPort().SetValue(1000)
	//tcp.DstPort().SetValue(1000)
	// // Configure custom bytes (hex string) in payload
	// cus.SetBytes(hex.EncodeToString([]byte("..QUICKSTART SNAPPI..")))

	// Optionally, print JSON representation of config
	if j, err := config.Marshal().ToJson(); err != nil {
		t.Fatal(err)
	} else {
		t.Log("Configuration: ", j)
	}

	// Push traffic configuration constructed so far to OTG
	if _, err := api.SetConfig(config); err != nil {
		t.Fatal(err)
	}

	// Start transmitting the packets from configured flow
	controlState := gosnappi.NewControlState()
	controlState.Traffic().FlowTransmit().
		SetState(gosnappi.StateTrafficFlowTransmitState.START).
		SetFlowNames([]string{flow.Name()})
	if _, err := api.SetControlState(controlState); err != nil {
		t.Fatal(err)
	}

	// t.Log(controlaction.Response().Protocol().Ipv4().Ping().Responses())
	// Fetch metrics for configured flow
	req := gosnappi.NewMetricsRequest()
	req.Flow().SetFlowNames([]string{flow.Name()})
	// and keep polling until either expectation is met or deadline exceeds
	deadline := time.Now().Add(30 * time.Second)
	for {
		metrics, err := api.GetMetrics(req)
		if err != nil {
			t.Fatalf("err = %v", err)
		}

		if time.Now().After(deadline) {
			t.Log("Timeout..........")
			break
		}

		// print YAML representation of flow metrics
		fmt.Println(metrics)

		// t.Log(metrics.FlowMetrics().Items()[0].Transmit())
		if metrics.FlowMetrics().Items()[0].Transmit() != gosnappi.FlowMetricTransmit.STARTED {
			t.Fatalf("Flow should be in state of started")
		}
		time.Sleep(10 * time.Second)
	}

	t.Log("Will Stop traffic...")
	controlState.Traffic().FlowTransmit().SetState(gosnappi.StateTrafficFlowTransmitState.STOP)
	if _, err := api.SetControlState(controlState); err != nil {
		t.Fatal(err)
	}
	metrics, err := api.GetMetrics(req)
	if err != nil {
		t.Fatalf("err = %v", err)
	}

	// print YAML representation of flow metrics
	fmt.Println(metrics)

	if metrics.FlowMetrics().Items()[0].Transmit() != gosnappi.FlowMetricTransmit.STOPPED {
		t.Fatalf("Flow should be in state of started")
	}
	fmt.Println("Test successful!")
}
