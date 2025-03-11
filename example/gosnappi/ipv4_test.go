package gosnappi_examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

// PORT1=10.109.121.181/1/1 PORT2=10.109.123.254/1/1 OTGSERVER=localhost:50051 go test -v -test.run TestIpv4

func TestIpv4(t *testing.T) {
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
	// udp := pkt.Add().Udp()
	cus := pkt.Add().Custom()

	eth.Dst().SetValue("00:11:22:33:44:55")
	//eth.Dst().Auto()
	//eth.Dst().SetValues([]string{"00:11:22:33:44:55", "00:11:22:33:44:56", "00:11:22:33:44:57"})
	//eth.Dst().Increment().SetStart("00:11:22:33:44:55").SetStep("00:00:00:00:00:01").SetCount(3)
	//eth.Dst().Decrement().SetStart("00:11:22:33:44:55").SetStep("00:00:00:00:00:01").SetCount(3)
	eth.Src().SetValue("00:11:22:33:44:66")
	//eth.Src().SetValues([]string{"00:11:22:33:44:66", "00:11:22:33:44:67", "00:11:22:33:44:68"})
	//eth.Src().Increment().SetStart("00:11:22:33:44:66").SetStep("00:00:00:00:00:01").SetCount(3)
	//eth.Src().Decrement().SetStart("00:11:22:33:44:66").SetStep("00:00:00:00:00:01").SetCount(3)
	eth.EtherType().SetValue(8914)
	//eth.EtherType().SetValues([]uint32{806, 804, 660, 661, 803})
	//eth.EtherType().Auto()
	//eth.EtherType().Increment().SetStart(600).SetStep(1).SetCount(20)
	//eth.EtherType().Decrement().SetStart(600).SetStep(1).SetCount(20)
	//eth.PfcQueue().SetValue(0)
	cus.SetBytes("123456")

	ipv4.Src().SetValue("10.1.1.1")
	//ipv4.Src().SetValues([]string{"10.0.0.1", "10.1.0.1", "10.2.0.1", "10.3.0.1"})
	//ipv4.Src().Decrement().SetStart("10.0.0.253").SetStep("0.0.0.4").SetCount(10)
	//ipv4.Src().Increment().SetStart("10.0.0.1").SetStep("0.0.0.2").SetCount(20)
	ipv4.Dst().SetValue("20.1.1.1")
	//ipv4.Dst().SetValues([]string{"20.0.0.1", "20.0.0.2", "20.0.0.3", "20.0.0.4"})
	//ipv4.Dst().Increment().SetStart("20.0.0.1").SetStep("0.0.0.2").SetCount(20)
	//ipv4.Dst().Decrement().SetStart("20.0.0.253").SetStep("0.0.0.4").SetCount(20)
	//ipv4.DontFragment().SetValues([]uint32{0, 1})
	//ipv4.DontFragment().SetValue(1)
	//ipv4.DontFragment().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.DontFragment().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.MoreFragments().SetValue(1)
	//ipv4.MoreFragments().SetValues([]uint32{0, 1})
	//ipv4.MoreFragments().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.MoreFragments().Decrement().SetStart(1).SetStep(1).SetCount(1)

	//ipv4.FragmentOffset().SetValue(31)
	//ipv4.FragmentOffset().SetValues([]uint32{31, 20, 10, 5})
	//ipv4.FragmentOffset().Decrement().SetStart(31).SetStep(4).SetCount(4)
	//ipv4.HeaderChecksum().SetGenerated("good")
	//ipv4.HeaderChecksum().SetGenerated("bad")
	//ipv4.HeaderChecksum().SetCustom(60000)
	//ipv4.HeaderLength().SetValues([]uint32{15, 12, 10})
	//ipv4.HeaderLength().Increment().SetStart(10).SetStep(1).SetCount(5)
	//ipv4.HeaderLength().Auto()
	//ipv4.HeaderLength().SetValue(15)
	//ipv4.Identification().SetValue(10)
	//ipv4.Identification().SetValues([]uint32{1, 10, 30, 65535})
	//ipv4.Identification().Increment().SetStart(1).SetStep(1).SetCount(3)
	//ipv4.Identification().Decrement().SetStart(10).SetStep(1).SetCount(5)
	//ipv4.Priority().Dscp().Ecn().SetValue(3)
	//ipv4.Priority().Dscp().Ecn().Increment().SetStart(0).SetStep(1).SetCount(3)
	//ipv4.Priority().Dscp().Ecn().Decrement().SetStart(3).SetStep(1).SetCount(3)
	//ipv4.Priority().Dscp().Ecn().SetValues([]uint32{0, 1, 2, 3})
	//ipv4.Priority().Tos().Delay().SetValue(1)
	//ipv4.Priority().Tos().Delay().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Delay().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Delay().SetValues([]uint32{0, 1})
	//ipv4.Priority().Tos().Monetary().SetValues([]uint32{0, 1})
	//ipv4.Priority().Tos().Monetary().SetValue(0)
	//ipv4.Priority().Tos().Monetary().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Monetary().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Reliability().SetValue(1)
	//ipv4.Priority().Tos().Reliability().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Reliability().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Reliability().SetValues([]uint32{0, 1})
	//ipv4.Priority().Tos().Throughput().SetValue(1)
	//ipv4.Priority().Tos().Throughput().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Throughput().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Throughput().SetValues([]uint32{0, 1})
	//ipv4.Priority().Tos().Unused().SetValue(1)
	//ipv4.Priority().Tos().Unused().Increment().SetStart(0).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Unused().Decrement().SetStart(1).SetStep(1).SetCount(1)
	//ipv4.Priority().Tos().Unused().SetValues([]uint32{0, 1})
	//ipv4.Priority().Dscp().Phb().SetValue(56)
	//ipv4.Priority().Dscp().Phb().SetValues([]uint32{1, 12, 20, 63})
	//ipv4.Priority().Dscp().Phb().Increment().SetStart(0).SetStep(10).SetCount(30)
	//ipv4.Priority().Dscp().Phb().Decrement().SetStart(30).SetStep(1).SetCount(10)
	//ipv4.Priority().Tos().Precedence().SetValue(2)
	//ipv4.Priority().Tos().Precedence().SetValues([]uint32{0, 1, 2, 3, 4, 5, 6, 7})
	//ipv4.Priority().Tos().Precedence().Increment().SetStart(0).SetStep(1).SetCount(7)
	//ipv4.Priority().Tos().Precedence().Decrement().SetStart(7).SetStep(1).SetCount(7)
	//ipv4.TimeToLive().Increment().SetStart(0).SetStep(1).SetCount(64)
	//ipv4.TimeToLive().Decrement().SetStart(255).SetStep(1).SetCount(60)
	//ipv4.TimeToLive().SetValues([]uint32{10, 20, 30, 255})
	//ipv4.TimeToLive().SetValue(255)
	//ipv4.TotalLength().SetValue(255)
	//ipv4.TotalLength().Increment().SetStart(1).SetStep(1).SetCount(100)
	//ipv4.TotalLength().Decrement().SetStart(65535).SetStep(1).SetCount(100)
	//ipv4.TotalLength().SetValues([]uint32{10, 20, 30, 50, 100})
	//ipv4.TotalLength().Auto()
	//ipv4.Version().SetValues([]uint32{4, 6})
	//ipv4.Version().Increment().SetStart(1).SetStep(1).SetCount(4)
	//ipv4.Version().Decrement().SetStart(4).SetStep(1).SetCount(4)
	//ipv4.Version().SetValue(4)
	//ipv4.Protocol().SetValue(101)
	//ipv4.Protocol().SetValues([]uint32{100, 101, 200, 250})
	//ipv4.Protocol().Increment().SetStart(1).SetStep(4).SetCount(10)
	//ipv4.Protocol().Decrement().SetStart(255).SetStep(8).SetCount(10)
	//ipv4.Protocol().Auto()   //error

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
