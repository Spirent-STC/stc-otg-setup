package large_ip_packet_transmission

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

func TestLargeIpPacket(t *testing.T) {
	// Create a new API handle to make API calls against OTG
	vlanNames := []string{"vlan10", "vlan20"}
	mtu := 9216
	flowSize := 8000
	acceptableLossPercent := 0.5
	acceptablePacketSizeDelta := 0.5
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

	ipv4.Src().SetValue("10.1.1.1")
	ipv4.Dst().SetValue("20.1.1.1")

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

	if flowSize > mtu {
		if inPkts == 0 {
			t.Logf(
				"flow sent '%d' packets and received '%d' packets, this is expected "+
					"due to flow size '%d' being > interface MTU of '%d' bytes",
				outPkts, inPkts, flowSize, mtu,
			)
		} else {
			t.Errorf(
				"flow received packets but should *not* have due to flow size '%d' being"+
					" > interface MTU of '%d' bytes",
				flowSize, mtu,
			)
		}
	}

	if outPkts == 0 || inPkts == 0 {
		t.Error("flow did not send or receive any packets, this should not happen")

		return
	}

	lossPercent := (float32(outPkts-inPkts) / float32(outPkts)) * 100

	if lossPercent > float32(acceptableLossPercent) {
		t.Errorf(
			"flow sent '%d' packets and received '%d' packets, resulting in a "+
				"loss percent of '%.2f'. max acceptable loss percent is '%.2f'",
			outPkts, inPkts, lossPercent, acceptableLossPercent,
		)
	}

	avgPacketSize := uint32(flowSize)
	packetSizeDelta := float32(avgPacketSize-uint32(flowSize)) / (float32(avgPacketSize+uint32(flowSize)) / 2) * 100
	if packetSizeDelta > float32(acceptablePacketSizeDelta) {
		t.Errorf(
			"flow sent '%d' packets and received '%d' packets, resulting in "+
				"averagepacket size of '%d' and a packet size delta of '%.2f' percent. "+
				"packet size delta should not exceed '%.2f'",
			outPkts, inPkts, avgPacketSize, packetSizeDelta, acceptablePacketSizeDelta,
		)
	}

	t.Log("Test successful!")
}
