package gosnappi_examples

import (
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
)

// PORT1=10.109.121.181/1/1 PORT2=10.109.123.254/1/1 OTGSERVER=localhost:50051 go test -v -test.run TestIcmpv6

func TestIcmpv6(t *testing.T) {
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
	ipv6 := pkt.Add().Ipv6()
	icmpv6 := pkt.Add().Icmpv6()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv6.Dst().SetValue("2000::1")
	ipv6.Src().SetValue("2001::1")

	icmpv6.Echo().Checksum().SetCustom(65535)
	//icmpv6.Echo().Checksum().SetGenerated("bad")
	icmpv6.Echo().Code().SetValue(0)
	//icmpv6.Echo().Code().SetValues([]uint32{0, 10, 100, 200, 255})
	//icmpv6.Echo().Code().Increment().SetStart(0).SetStep(1).SetCount(254)
	//icmpv6.Echo().Code().Decrement().SetStart(255).SetStep(1).SetCount(254)
	icmpv6.Echo().Identifier().SetValue(65535)
	//icmpv6.Echo().Identifier().SetValues([]uint32{0, 100, 100, 100, 1000, 10000, 65535})
	//icmpv6.Echo().Identifier().Increment().SetStart(0).SetStep(1).SetCount(1000)
	//icmpv6.Echo().Identifier().Decrement().SetStart(65535).SetStep(1).SetCount(1000)
	icmpv6.Echo().SequenceNumber().SetValue(65535)
	//icmpv6.Echo().SequenceNumber().SetValues([]uint32{0, 100, 1000, 20000, 65535})
	//icmpv6.Echo().SequenceNumber().Increment().SetStart(0).SetStep(1).SetCount(1000)
	//icmpv6.Echo().SequenceNumber().Decrement().SetStart(65535).SetStep(1).SetCount(1000)
	//icmpv6.Echo().Type().SetValue(129)
	//icmpv6.Echo().Type().SetValues([]uint32{0, 10, 100, 200, 255})
	//icmpv6.Echo().Type().Increment().SetStart(0).SetStep(1).SetCount(10)
	//icmpv6.Echo().Type().Decrement().SetStart(255).SetStep(1).SetCount(100)

	fmt.Println("Test Gosnappi begin :")

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
