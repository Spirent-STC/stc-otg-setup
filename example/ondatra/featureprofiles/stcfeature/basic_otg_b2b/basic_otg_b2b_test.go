package basic_otg_b2b

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
)

type intf struct {
	// Name is the name of the interface.
	Name string
	// MAC is the MAC address for the interface.
	MAC string
}

var (
	// ateSrc describes the configuration parameters for the ATE port sourcing
	// a flow.
	ateSrc = &intf{
		Name: "port1",
		MAC:  "02:00:01:01:01:01",
	}

	ateDst = &intf{
		Name: "port2",
		MAC:  "02:00:02:01:01:01",
	}
)

// The testbed consists of ate:port1 -> dut:port1.
func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

func TestBasicOtgB2b(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	vlanNames := []string{"vlan10", "vlan20"}
	mtu := 9216
	flowSize := 8000
	// acceptableLossPercent := 0.5
	// acceptablePacketSizeDelta := 0.5
	vlanId := []uint32{10, 20}
	ate := ondatra.ATE(t, "ate")

	otg := ate.OTG()
	//topology := otg.NewConfig(t)
	topology := gosnappi.NewConfig()
	for i, p := range []*intf{ateSrc, ateDst} {
		topology.Ports().Add().SetName(p.Name)
		dev := topology.Devices().Add().SetName(p.Name)
		eth := dev.Ethernets().Add().SetName(fmt.Sprintf("%s_ETH", p.Name)).SetMac(p.MAC).SetMtu(uint32(mtu))
		eth.Vlans().Add().SetName(vlanNames[i]).SetId(vlanId[i]).SetPriority(2).SetTpid("x9300")
		eth.Connection().SetPortName(dev.Name())
	}

	// Configure a flow and set previously created test port as one of endpoints
	flow := topology.Flows().Add().SetName("flow1")
	flow.TxRx().Port().SetTxName(ateSrc.Name).SetRxNames([]string{ateDst.Name})
	// and enable tracking flow metrics
	flow.Metrics().SetEnable(true)

	// Configure number of packets to transmit for previously configured flow
	flow.Duration().Continuous()
	// and fixed byte size of all packets in the flow
	flow.Size().SetFixed(uint32(flowSize))
	flow.Rate().SetPercentage(1.0)

	// Configure protocol headers for all packets in the flow
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv4 := pkt.Add().Ipv4()

	eth.Dst().SetValue("00:11:22:33:44:55")
	eth.Src().SetValue("00:11:22:33:44:66")

	ipv4.Src().Increment().SetStart("04")
	ipv4.Src().Increment().SetStep("01")
	ipv4.Src().Increment().SetCount(1)
	ipv4.Src().SetValue("10.1.1.1")

	ipv4.Dst().Increment().SetStart("04")
	ipv4.Dst().Increment().SetStep("01")
	ipv4.Dst().Increment().SetCount(1)
	ipv4.Dst().SetValue("20.1.1.1")

	ipv4.HeaderChecksum().SetGenerated("Generated.Enum.good")
	ipv4.HeaderChecksum().SetCustom(1)

	otg.PushConfig(t, topology)

	t.Logf("Starting traffic...")
	otg.StartTraffic(t)

	sleep_seconds := 10 * time.Second
	t.Logf("Sleeping for %s ...Hold the traffic to check TestIQ result", sleep_seconds)
	time.Sleep(sleep_seconds)

	t.Logf("Stopping traffic...")
	otg.StopTraffic(t)

	// Avoid a race with telemetry updates.
	sleep_seconds = 20 * time.Second
	t.Logf("Sleeping for %s to wait telemetry update ...", sleep_seconds)

	t.Logf("Get flow stats ...")
	metrics := gnmi.Get(t, otg, gnmi.OTG().Flow("flow1").State())

	b, err := json.Marshal(metrics)
	if err == nil {
		t.Logf("Metrics: %+v", string(b))
	} else {
		t.Logf("Metrics: %+v", metrics)
	}
	outPkts := metrics.GetCounters().GetOutPkts()
	inPkts := metrics.GetCounters().GetInPkts()
	// inOctets := metrics.GetCounters().GetInOctets()

	t.Log(outPkts)
	t.Log(inPkts)
	t.Log(metrics)
	if outPkts != inPkts {
		t.Fatalf("Flow should be in state of started")
	}

	t.Log("Test successful!")
}

func TestBasicOtgB2bV6(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	flowSize := 8000
	ate := ondatra.ATE(t, "ate")

	otg := ate.OTG()
	//topology := otg.NewConfig(t)
	topology := gosnappi.NewConfig()

	ptx := topology.Ports().Add().SetName("port1").SetLocation("//10.61.37.121/1/1")
	prx := topology.Ports().Add().SetName("port2").SetLocation("//10.61.37.121/1/2")
	dev_tx := topology.Devices().Add().SetName(fmt.Sprintf("%s_DEV", ptx.Name()))
	eth_tx := dev_tx.Ethernets().Add().SetName(fmt.Sprintf("%s_ETH", ptx.Name())).SetMac("00:10:94:00:00:02")
	eth_tx.Connection().SetPortName(ptx.Name())
	// eth_tx.SetPortName(ptx.Name())

	dev_rx := topology.Devices().Add().SetName(fmt.Sprintf("%s_DEV", prx.Name()))
	eth_rx := dev_rx.Ethernets().Add().SetName(fmt.Sprintf("%s_ETH", prx.Name())).SetMac("00:10:94:00:00:03")
	eth_rx.Connection().SetPortName(prx.Name())
	// eth_rx.SetPortName(prx.Name())

	// Configure a flow and set previously created test port as one of endpoints
	flow := topology.Flows().Add().SetName("flow1")
	flow.TxRx().Port().SetTxName(ateSrc.Name).SetRxNames([]string{ateDst.Name})
	// and enable tracking flow metrics
	flow.Metrics().SetEnable(true)

	// Configure number of packets to transmit for previously configured flow
	flow.Duration().Continuous()
	// and fixed byte size of all packets in the flow
	flow.Size().SetFixed(uint32(flowSize))
	flow.Rate().SetPercentage(1.0)

	// Configure protocol headers for all packets in the flow
	pkt := flow.Packet()
	eth := pkt.Add().Ethernet()
	ipv6 := pkt.Add().Ipv6()

	eth.Dst().SetValue("00:10:94:00:00:02")
	eth.Src().SetValue("00:10:94:00:00:03")

	ipv6.Src().Increment().SetStart("04")
	ipv6.Src().Increment().SetStep("01")
	ipv6.Src().Increment().SetCount(1)
	ipv6.Src().SetValue("2001::2/64")

	ipv6.Dst().Increment().SetStart("04")
	ipv6.Dst().Increment().SetStep("01")
	ipv6.Dst().Increment().SetCount(1)
	ipv6.Dst().SetValue("2001::3/64")

	otg.PushConfig(t, topology)

	t.Logf("Starting traffic...")
	otg.StartTraffic(t)

	sleep_seconds := 10 * time.Second
	t.Logf("Sleeping for %s ...Hold the traffic to check TestIQ result", sleep_seconds)
	time.Sleep(sleep_seconds)

	t.Logf("Stopping traffic...")
	otg.StopTraffic(t)

	// Avoid a race with telemetry updates.
	sleep_seconds = 20 * time.Second
	t.Logf("Sleeping for %s to wait telemetry update ...", sleep_seconds)

	t.Logf("Get flow stats ...")
	metrics := gnmi.Get(t, otg, gnmi.OTG().Flow("flow1").State())

	b, err := json.Marshal(metrics)
	if err == nil {
		t.Logf("Metrics: %+v", string(b))
	} else {
		t.Logf("Metrics: %+v", metrics)
	}
	outPkts := metrics.GetCounters().GetOutPkts()
	inPkts := metrics.GetCounters().GetInPkts()
	// inOctets := metrics.GetCounters().GetInOctets()

	t.Log(outPkts)
	t.Log(inPkts)
	t.Log(metrics)
	if outPkts != inPkts {
		t.Fatalf("Flow should be in state of started")
	}

	t.Log("Test successful!")
}
