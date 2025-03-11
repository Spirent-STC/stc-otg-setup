package gosnappi_examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

// PORT1=10.109.121.181/1/1 PORT2=10.109.123.254/1/1 OTGSERVER=localhost:50051 go test -v -test.run TestGre

func TestGre(t *testing.T) {
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
	gre := pkt.Add().Gre()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv4.Src().SetValue("10.1.1.1")
	ipv4.Dst().SetValue("20.1.1.1")

	gre.ChecksumPresent().SetValue(1)
	//gre.Checksum().SetCustom(10)
	gre.Reserved1().SetValue(1000)
	//gre.ChecksumPresent().SetValues([]uint32{0, 1})
	//gre.ChecksumPresent().Increment().SetStart(0).SetStep(1).SetCount(1)
	//gre.ChecksumPresent().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//gre.Protocol().SetValue(65535)
	//gre.Protocol().SetValues([]uint32{9999})
	//gre.Protocol().Increment().SetStart(800).SetStep(1).SetCount(10)
	//gre.Protocol().Decrement().SetStart(800).SetStep(1).SetCount(10)
	gre.Reserved0().SetValue(511)
	//gre.Reserved0().SetValues([]uint32{0, 10, 100, 511})
	//gre.Reserved0().Increment().SetStart(0).SetStep(1).SetCount(510)
	//gre.Reserved0().Decrement().SetStart(511).SetStep(1).SetCount(510)
	gre.Version().SetValue(0)
	//gre.Version().SetValues([]uint32{0, 1, 2, 3, 4, 5, 6, 7})
	//gre.Version().Increment().SetStart(0).SetStep(1).SetCount(7)
	//gre.Version().Decrement().SetStart(7).SetStep(1).SetCount(7)

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
