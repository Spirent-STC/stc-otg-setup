// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rt_5_2_aggregate_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/featureprofiles/internal/attrs"
	"github.com/openconfig/featureprofiles/internal/fptest"
	"github.com/openconfig/featureprofiles/internal/otgutils"
	"github.com/openconfig/ondatra"
	"github.com/openconfig/ondatra/gnmi"
	"github.com/openconfig/ondatra/gnmi/oc"
	otgtelemetry "github.com/openconfig/ondatra/gnmi/otg"
	"github.com/openconfig/ygnmi/ygnmi"
)

func TestMain(m *testing.M) {
	fptest.RunTests(m)
}

// Settings for configuring the aggregate testbed with the test
// topology.  IxNetwork flow requires both source and destination
// networks be configured on the ATE.  It is not possible to send
// packets to the ether.
//
// The testbed consists of ate:port1 -> dut:port1 and dut:port{2-9} ->
// ate:port{2-9}.  The first pair is called the "source" pair, and the
// second aggregate link the "destination" pair.
//
//   - Source: ate:port1 -> dut:port1 subnet 192.0.2.0/30 2001:db8::0/126
//   - Destination: dut:port{2-9} -> ate:port{2-9}
//     subnet 192.0.2.4/30 2001:db8::4/126
//
// Note that the first (.0, .4) and last (.3, .7) IPv4 addresses are
// reserved from the subnet for broadcast, so a /30 leaves exactly 2
// usable addresses.  This does not apply to IPv6 which allows /127
// for point to point links, but we use /126 so the numbering is
// consistent with IPv4.
//
// A traffic flow is configured from ate:port1 as source and ate:port{2-9}
// as destination.
const (
	plen4 = 24
	plen6 = 126
)

var (
	dutSrc = attrs.Attributes{
		Name:    "dutsrc",
		MAC:     "02:22:01:00:00:01",
		IPv4:    "192.0.2.2",
		IPv6:    "2001:db8::2",
		IPv4Len: plen4,
		IPv6Len: plen6,
	}

	ateDst = attrs.Attributes{
		Name:    "atedst",
		MAC:     "02:12:01:00:00:01",
		IPv4:    "192.0.2.6",
		IPv6:    "2001:db8::6",
		IPv4Len: plen4,
		IPv6Len: plen6,
	}

	expectedStatus = make(map[string]any)
	breakLinkIds   = []uint32{}
)

const (
	lagTypeLACP   = oc.IfAggregate_AggregationType_LACP
	lagTypeSTATIC = oc.IfAggregate_AggregationType_STATIC
)

type testCase struct {
	lagType oc.E_IfAggregate_AggregationType

	ate *ondatra.ATEDevice
	top gosnappi.Config

	// atePorts is the set of ports on the ATE -- the first, as with the DUT
	// is not configured in the aggregate interface.
	// is not configured in the aggregate interface.
	atePorts []*ondatra.Port
	aggID1   string
	aggID2   string
}

func (tc *testCase) configureATE(t *testing.T) {
	tc.configureATE1(t)
	tc.configureATE2(t)

	t.Log(tc.top.Marshal().ToJson())

	tc.ate.OTG().PushConfig(t, tc.top)
	tc.ate.OTG().StartProtocols(t)
}

func (tc *testCase) configureATE1(t *testing.T) {
	// Adding the rest of the ports to the configuration and to the LAG
	agg := tc.top.Lags().Add().SetName(ateDst.Name)
	if tc.lagType == lagTypeSTATIC {
		lagId, _ := strconv.Atoi(tc.aggID1)
		agg.Protocol().Static().SetLagId(uint32(lagId))
		for i, p := range tc.atePorts[0:4] {
			port := tc.top.Ports().Add().SetName(p.ID())
			newMac, err := incrementMAC(ateDst.MAC, i+1)
			if err != nil {
				t.Fatal(err)
			}
			agg.Ports().Add().SetPortName(port.Name()).Ethernet().SetMac(newMac).SetName("LAGRx-" + strconv.Itoa(i))
		}
	} else {
		agg.Protocol().Lacp().SetActorKey(1).SetActorSystemPriority(1).SetActorSystemId(ateDst.MAC)
		for i, p := range tc.atePorts[0:4] {
			port := tc.top.Ports().Add().SetName(p.ID())
			newMac, err := incrementMAC(ateDst.MAC, i+1)
			if err != nil {
				t.Fatal(err)
			}
			lagPort := agg.Ports().Add().SetPortName(port.Name())
			lagPort.Ethernet().SetMac(newMac).SetName("LAGRx-" + strconv.Itoa(i))
			lagPort.Lacp().SetActorActivity("active").SetActorPortNumber(uint32(i) + 1).SetActorPortPriority(1).SetLacpduTimeout(0)
		}
	}

	dstDev := tc.top.Devices().Add().SetName(agg.Name() + ".dev")
	dstEth := dstDev.Ethernets().Add().SetName(ateDst.Name + ".Eth").SetMac(ateDst.MAC)
	dstEth.Connection().SetLagName(agg.Name())
	dstEth.Ipv4Addresses().Add().SetName(ateDst.Name + ".IPv4").SetAddress(ateDst.IPv4).SetGateway(dutSrc.IPv4).SetPrefix(uint32(ateDst.IPv4Len))
	dstEth.Ipv6Addresses().Add().SetName(ateDst.Name + ".IPv6").SetAddress(ateDst.IPv6).SetGateway(dutSrc.IPv6).SetPrefix(uint32(ateDst.IPv6Len))
}

// simulate DUT as the flow source
func (tc *testCase) configureATE2(t *testing.T) {

	// Adding the rest of the ports to the configuration and to the LAG
	agg := tc.top.Lags().Add().SetName(dutSrc.Name)
	if tc.lagType == lagTypeSTATIC {
		lagId, _ := strconv.Atoi(tc.aggID2)
		agg.Protocol().Static().SetLagId(uint32(lagId))
		for i, p := range tc.atePorts[4:] {
			port := tc.top.Ports().Add().SetName(p.ID())
			newMac, err := incrementMAC(dutSrc.MAC, i+1)
			if err != nil {
				t.Fatal(err)
			}
			agg.Ports().Add().SetPortName(port.Name()).Ethernet().SetMac(newMac).SetName("LAGRx-" + strconv.Itoa(i))
		}
	} else {
		agg.Protocol().Lacp().SetActorKey(1).SetActorSystemPriority(1).SetActorSystemId(dutSrc.MAC)
		for i, p := range tc.atePorts[4:] {
			port := tc.top.Ports().Add().SetName(p.ID())
			newMac, err := incrementMAC(dutSrc.MAC, i+1)
			if err != nil {
				t.Fatal(err)
			}
			lagPort := agg.Ports().Add().SetPortName(port.Name())
			lagPort.Ethernet().SetMac(newMac).SetName("LAGRx-" + strconv.Itoa(i))
			lagPort.Lacp().SetActorActivity("active").SetActorPortNumber(uint32(i) + 1).SetActorPortPriority(1).SetLacpduTimeout(0)
		}
	}

	// tc.top.Flows().Clear().Items()

	// device 1 and flow 1
	for i := range []int{0, 1, 2, 3} {
		devname := fmt.Sprintf("%s.dev%d", agg.Name(), i)
		srcDev := tc.top.Devices().Add().SetName(devname)

		ethname := fmt.Sprintf("%s.Eth%d", dutSrc.Name, i)
		dutSrcMac, _ := incrementMAC(dutSrc.MAC, i)
		srcEth := srcDev.Ethernets().Add().SetName(ethname).SetMac(dutSrcMac)
		srcEth.Connection().SetLagName(agg.Name())

		v4name := fmt.Sprintf("%s.IPv4%d", dutSrc.Name, i)
		v6name := fmt.Sprintf("%s.IPv6%d", dutSrc.Name, i)
		dutSrcIPv4 := incrementIPV4Address(dutSrc.IPv4, i)
		dutSrcIPv6 := incrementIPV6Address(t, dutSrc.IPv6, i, "host")
		srcEth.Ipv4Addresses().Add().SetName(v4name).SetAddress(dutSrcIPv4).SetGateway(ateDst.IPv4).SetPrefix(uint32(dutSrc.IPv4Len))
		srcEth.Ipv6Addresses().Add().SetName(v6name).SetAddress(dutSrcIPv6).SetGateway(ateDst.IPv6).SetPrefix(uint32(dutSrc.IPv6Len))

		flowname := fmt.Sprintf("flow%d", i)
		flow := tc.top.Flows().Add().SetName(flowname)
		flow.Metrics().SetEnable(true)
		flow.Size().SetFixed(128)
		flow.Packet().Add().Ethernet().Src().SetValue(dutSrcMac)

		flow.TxRx().Device().SetTxNames([]string{v4name}).SetRxNames([]string{ateDst.Name + ".IPv4"})
		v4 := flow.Packet().Add().Ipv4()
		v4.Src().SetValue(dutSrcIPv4)
		v4.Dst().SetValue(ateDst.IPv4)
	}

}

func (tc *testCase) verifyATE(t *testing.T) {
	// State for the interface.
	time.Sleep(3 * time.Second)
	otgutils.LogLAGMetrics(t, tc.ate.OTG(), tc.top)

	if tc.lagType == oc.IfAggregate_AggregationType_LACP {
		otgutils.LogLACPMetrics(t, tc.ate.OTG(), tc.top)
	}

	tc.verifyATE1(t)
	tc.verifyATE2(t)

}

func (tc *testCase) verifyLoadBalance(t *testing.T) {

	beforeTrafficCounters := tc.getCounters(t, "before")

	tc.ate.OTG().StartTraffic(t)
	time.Sleep(15 * time.Second)
	tc.ate.OTG().StopTraffic(t)

	otgutils.LogPortMetrics(t, tc.ate.OTG(), tc.top)
	otgutils.LogFlowMetrics(t, tc.ate.OTG(), tc.top)
	otgutils.LogLAGMetrics(t, tc.ate.OTG(), tc.top)

	for i := range []int{0, 1, 2, 3} {
		flowname := fmt.Sprintf("flow%d", i)
		recvMetric := gnmi.Get(t, tc.ate.OTG(), gnmi.OTG().Flow(flowname).State())
		pkts := recvMetric.GetCounters().GetOutPkts()

		if pkts == 0 {
			t.Errorf("Flow sent packets: got %v, want non zero", pkts)
		} else {
			t.Logf("Flow sent packets - %v", pkts)
		}
	}

	afterTrafficCounters := tc.getCounters(t, "after")
	tc.verifyCounterDiff(t, beforeTrafficCounters, afterTrafficCounters)

}

func (tc *testCase) verifyPortOperStatus(t *testing.T, ap *ondatra.Port) {
	var expect = expectedStatus[ap.Name()].(otgtelemetry.E_Port_Link)
	t.Logf("Checking oper-status for port %s expect is %v", ap.Name(), expect)
	portMetrics := gnmi.Get(t, tc.ate.OTG(), gnmi.OTG().Port(ap.ID()).State())
	if portMetrics.GetLink() != expect {
		t.Errorf("%s oper-status got %v, want %v", ap.ID(), portMetrics.GetLink(), expect)
	}
}

func (tc *testCase) verifyLAGOperStatus(t *testing.T, LagName string) {
	var expect = expectedStatus[LagName].(otgtelemetry.E_Lag_OperStatus)
	t.Logf("Checking oper-status for LAG %s expect is %s", LagName, expect)
	gnmi.Watch(t, tc.ate.OTG(), gnmi.OTG().Lag(LagName).OperStatus().State(), time.Minute, func(val *ygnmi.Value[otgtelemetry.E_Lag_OperStatus]) bool {
		state, present := val.Val()
		return present && state == expect
	}).Await(t)
}

func (tc *testCase) verifyATE1(t *testing.T) {
	for _, p := range tc.atePorts[0:4] {
		tc.verifyPortOperStatus(t, p)
	}
	tc.verifyLAGOperStatus(t, ateDst.Name)
}

func (tc *testCase) verifyATE2(t *testing.T) {
	for _, p := range tc.atePorts[4:] {
		tc.verifyPortOperStatus(t, p)
	}
	tc.verifyLAGOperStatus(t, dutSrc.Name)
}

// sortPorts sorts the ports by the testbed port ID.
func sortPorts(ports []*ondatra.Port) []*ondatra.Port {
	sort.SliceStable(ports, func(i, j int) bool {
		return ports[i].ID() < ports[j].ID()
	})
	return ports
}

// incrementMAC increments the MAC by i. Returns error if the mac cannot be parsed or overflows the mac address space
func incrementMAC(mac string, i int) (string, error) {
	macAddr, err := net.ParseMAC(mac)
	if err != nil {
		return "", err
	}
	convMac := binary.BigEndian.Uint64(append([]byte{0, 0}, macAddr...))
	convMac = convMac + uint64(i)
	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.BigEndian, convMac)
	if err != nil {
		return "", err
	}
	newMac := net.HardwareAddr(buf.Bytes()[2:8])
	return newMac.String(), nil
}

// incrementIPV4Address increments the IPv4 address by i
func incrementIPV4Address(ip string, i int) string {
	ipAddr := net.ParseIP(ip)
	convIP := binary.BigEndian.Uint32(ipAddr.To4())
	convIP = convIP + uint32(i)
	newIP := make(net.IP, 4)
	binary.BigEndian.PutUint32(newIP, convIP)
	return newIP.String()
}

func incrementIPV6Address(t *testing.T, address string, i int, part string) string {
	addr := net.ParseIP(address)
	var oct int
	switch part {
	case "network":
		oct = 13
	case "host":
		oct = 15
	default:
		t.Errorf("Invalid value")
	}
	for j := 0; j < i; j++ {
		addr[oct]++
	}
	return addr.String()
}

func (tc *testCase) breakLinks(t *testing.T) {
	for i := range breakLinkIds {
		port := tc.atePorts[i]
		portStateAction := gosnappi.NewControlState()
		portStateAction.Port().Link().SetPortNames([]string{port.ID()}).SetState(gosnappi.StatePortLinkState.DOWN)
		tc.ate.OTG().SetControlState(t, portStateAction)
	}
}

// normalize normalizes the input values so that the output values sum
// to 1.0 but reflect the proportions of the input.  For example,
// input [1, 2, 3, 4] is normalized to [0.1, 0.2, 0.3, 0.4].
func normalize(xs []uint64) (ys []float64, sum uint64) {
	for _, x := range xs {
		sum += x
	}
	ys = make([]float64, len(xs))
	for i, x := range xs {
		ys[i] = float64(x) / float64(sum)
	}
	return ys, sum
}

var approxOpt = cmpopts.EquateApprox(0 /* frac */, 0.01 /* absolute */)

// portWants converts the nextHop wanted weights to per-port wanted
// weights listed in the same order as atePorts.
func (tc *testCase) portWants() []float64 {
	numPorts := len(tc.atePorts[4:])
	weights := []float64{}
	for i := 0; i < numPorts; i++ {
		weights = append(weights, 1/float64(numPorts))
	}
	return weights
}

func (tc *testCase) verifyCounterDiff(t *testing.T, before, after map[string]*otgtelemetry.Port_Counters) {
	b := &strings.Builder{}
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)

	fmt.Fprint(w, "Port Counter Deltas\n\n")
	fmt.Fprint(w, "Name\tInPkts\tInOctets\tOutPkts\tOutOctets\n")
	allInPkts := []uint64{}
	allOutPkts := []uint64{}

	for port := range before {
		inPkts := after[port].GetInFrames() - before[port].GetInFrames()
		allInPkts = append(allInPkts, inPkts)
		inOctets := after[port].GetInOctets() - before[port].GetInOctets()
		outPkts := after[port].GetOutFrames() - before[port].GetOutFrames()
		allOutPkts = append(allOutPkts, outPkts)
		outOctets := after[port].GetOutOctets() - before[port].GetOutOctets()

		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\n",
			port,
			inPkts, inOctets,
			outPkts, outOctets)
	}
	got, outSum := normalize(allOutPkts)
	want := tc.portWants()
	t.Logf("outPkts normalized got: %v", got)
	t.Logf("want: %v", want)
	t.Run("Ratio", func(t *testing.T) {
		if diff := cmp.Diff(want, got, approxOpt); diff != "" {
			t.Errorf("Packet distribution ratios -want,+got:\n%s", diff)
		}
	})
	t.Run("Loss", func(t *testing.T) {
		if allInPkts[0] > outSum {
			t.Errorf("Traffic flow received %d packets, sent only %d",
				allOutPkts[0], outSum)
		}
	})
	w.Flush()

	t.Log(b)
}

func (tc *testCase) getCounters(t *testing.T, when string) map[string]*otgtelemetry.Port_Counters {
	results := make(map[string]*otgtelemetry.Port_Counters)
	b := &strings.Builder{}
	w := tabwriter.NewWriter(b, 0, 0, 1, ' ', 0)

	fmt.Fprint(w, "Raw Port Counters\n\n")
	fmt.Fprint(w, "Name\tInPkts\tInOctets\tOutPkts\tOutOctets\n")
	for _, port := range tc.atePorts[4:] {
		state := gnmi.Get(t, tc.ate.OTG(), gnmi.OTG().Port(port.ID()).State())
		counters := state.GetCounters()
		results[port.Name()] = counters
		fmt.Fprintf(w, "%s\t%d\t%d\t%d\t%d\n",
			port.Name(),
			counters.GetInFrames(), counters.GetInOctets(),
			counters.GetOutFrames(), counters.GetOutOctets())
	}
	w.Flush()

	t.Log(b)

	return results
}

func TestNegotiation(t *testing.T) {
	ate := ondatra.ATE(t, "ate")

	lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeSTATIC, lagTypeLACP}
	// lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeSTATIC}
	// lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeLACP}

	for _, lagType := range lagTypes {

		top := gosnappi.NewConfig()
		// Clean otg with an empty config
		ate.OTG().PushConfig(t, top)

		tc := &testCase{
			ate:      ate,
			top:      top,
			lagType:  lagType,
			atePorts: sortPorts(ate.Ports()),
			aggID1:   "001",
			aggID2:   "002",
		}
		t.Run(fmt.Sprintf("LagType=%s", lagType), func(t *testing.T) {
			tc.configureATE(t)

			for _, p := range tc.atePorts {
				expectedStatus[p.Name()] = otgtelemetry.Port_Link_UP
			}
			expectedStatus[ateDst.Name] = otgtelemetry.Lag_OperStatus_UP
			expectedStatus[dutSrc.Name] = otgtelemetry.Lag_OperStatus_UP

			t.Run("VerifyATE all up", tc.verifyATE)

			//////////////////////////////////////////////////////////////
			indexBreakLink := uint32(0)
			breakLinkIds = []uint32{indexBreakLink}
			t.Run("Break 1 link", tc.breakLinks)
			expectedStatus[tc.atePorts[indexBreakLink].Name()] = otgtelemetry.Port_Link_DOWN
			expectedStatus[tc.atePorts[indexBreakLink+4].Name()] = otgtelemetry.Port_Link_DOWN

			t.Run("VerifyATE 1 link broken", tc.verifyATE)

			//////////////////////////////////////////////////////////////
			breakLinkIds = []uint32{0, 1, 2, 3, 4, 5, 6, 7}
			t.Run("Break all links", tc.breakLinks)
			for _, p := range tc.atePorts {
				expectedStatus[p.Name()] = otgtelemetry.Port_Link_DOWN
			}
			expectedStatus[ateDst.Name] = otgtelemetry.Lag_OperStatus_DOWN
			expectedStatus[dutSrc.Name] = otgtelemetry.Lag_OperStatus_DOWN

			t.Run("VerifyATE all links broken", tc.verifyATE)
		})
	}
}

func TestTraffic(t *testing.T) {
	ate := ondatra.ATE(t, "ate")

	// lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeSTATIC, lagTypeLACP}
	// lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeSTATIC}
	lagTypes := []oc.E_IfAggregate_AggregationType{lagTypeLACP}

	for _, lagType := range lagTypes {

		top := gosnappi.NewConfig()
		// Clean otg with an empty config
		ate.OTG().PushConfig(t, top)

		tc := &testCase{
			ate:      ate,
			top:      top,
			lagType:  lagType,
			atePorts: sortPorts(ate.Ports()),
			aggID1:   "001",
			aggID2:   "002",
		}
		t.Run(fmt.Sprintf("LagType=%s", lagType), func(t *testing.T) {
			tc.configureATE(t)

			for _, p := range tc.atePorts {
				expectedStatus[p.Name()] = otgtelemetry.Port_Link_UP
			}
			expectedStatus[ateDst.Name] = otgtelemetry.Lag_OperStatus_UP
			expectedStatus[dutSrc.Name] = otgtelemetry.Lag_OperStatus_UP

			t.Run("VerifyATE all up", tc.verifyATE)

			t.Run("VerifyATE Load Balance", tc.verifyLoadBalance)

		})
	}
}
