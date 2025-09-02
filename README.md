# About Spycraft

Spycraft can be used to provide standardized call accounting and call analysis
for external SIP services such as an IP-PBX, proxies, boarder controllers, and
SIP enabled gateways. It does so by collecting from external packet inspection.
This makes it easy to integrate standardized call accounting with any existing
SIP switching system without having to develop and integrate such features
internally in such products.

Part of what makes spycraft special is zero-copy parsing of captured packets
and the use of pipelining. This and BPF rule generation greatly reduces system
overhead and will make it very possible to run spycraft as a minimal network
appliance on a SBC even for very large call centers or cloud hosted services.

Spycraft is also usable as a toolset for SIP call analysis and can be tested
and verified for an existing switching platform using pcap captures of network
traffic. This also makes it possible to create repeatable integration test
cases for spycraft out of such capture files. Spycraft will also be capable of
creating its own pcap captures.

As a network collection daemon, sipcraft can operate unprivileged on a local
machine co-resident with your SIP service, as a privileged promiscuous monitor
appliance monitoring call traffic and remote devices on a network, or even as a
non-privileged process on a routers network traffic mirror port. It can also
locally be used to monitor and produce call reports if co-resident with a
secure IP-PBX operating over a WireGuard network such as tailscale.

Spycraft is meant to integrate with external call-accounting systems, including
its own that will use postgres, and will eventually include it's own call
accounting service that will use that. In theory it should be possible to
integrate collected call records from SIP call analysis in things like Radius
as well.

The initial release was meant to confidently demonstrate the basic concept and
capabilities of what a generic sip network monitoring based call analysis
engine like spycraft can potentially do as well as what role it can potentially
play in supporting cloud based telephony and IP-PBX deployments. Could
spycraft also eventually be adapted for VoIP intrusion detection and real-time
call traffic monitoring or alerting?  Certainly.

## Dependencies

Spycraft uses Go-Packet, which is itself a wrapper around libpcap, to do
network monitoring and access pcap files. This means that a valid version of
libpcap must be available in your build environment or, if cross-compiling,
your sysroot, as Go-Packet uses cgo to call libpcap directly.

## Licensing

I have chosen to use the Affero General Public License for this package. I may
make this software available under alternative licensing terms to specific
entities, or may introduce formal dual licensing. But at this point, I have no
specific intent other than to maximally protect the freedom of those who may
encounter this package, including end users who may not otherwise be given
rights if the application is ran on their network by a commercial entity who
has no obligation or interest to formally convey the software to them. If I
later offer selective re-license I will formally introduce dual-liancing.

While spycraft is initially a network ``consumption'' utility it will
eventually include a web service for things like call monitoring and call
accounting. It likely will come to feature a web api, too. This also suggests
strong reasons for chosing to use the AGPL rather than the regular GPL.

## Participation

This project is offered as free (as in freedom) software for public use and has
a public project page at https://www.github.com/dyfet/spycraft which has an
issue tracker where you can submit public bug reports and a public git
repository. Patches and merge requests may be submitted in the issue tracker or
thru email. Other details about participation may be found in
CONTRIBUTING.md.

## Testing

Testing offline with .pcap files is very easy. If you want to test capture from
your build environment you can run "make setcap" to give permissions to your
debug build. You can then run the target/debug/spycraft executable in capture
node. If you want to test promiscuous mode you may need to test as root.

