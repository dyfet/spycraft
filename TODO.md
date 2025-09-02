# Roadmap

This is meant to give an idea and roadmap of where future development may end
up going. This is just a very early release that simply and effectively
demonstrates a full proof of concept and the potential for doing network
observability tooling like spycraft in go.

## Finish wiring up call states and collation

There is still a lot of work to do in call state event handling to generate and
track the state of externally observed sip sessions. I am also not yet sure how
far I will get before I do an initial release. I initially did a very crude
prototype that already did not separately separate inbound / outbound
requests. It also lacked collation for tracking b2bua servers, which many of
mine do, where participating call legs have independent call id's.

## Completing call processing model and basic CDR logging

This would be an initial mvp release point for spycraft. It should be able to
write call leg records to a database. Call histories can happen by merging
multiple call legs under a collation id and node name. This would allow
multopile spycraft instances, each tracking a different interface, to produce
a single complete call query. The base of a complete call query would be the
initiating incoming call leg wherever it happens.

## TCP packet assembly and sip TCP

There is some initial work on a TCP packet re-assembler already in byteshark.

## Media capture

A separate subnet packet filter session for spycraft to capture realtime media.
If used in promiscuous mode it may be able to capture media sessions even from
p2p media that is not server routed. This might be used to meet legal
obligations for call recording or to do automated transcriptions.

## Call detail reporting

Beyond generating a simple CDR text record for completed calls, there is a
desire to inject radius records, and a postgresql database back-end. This
backend could then be serviced by a web interface as a complete call account
and media retrieval system.

## Alerting and intrusion detection

This would involve using the real-time capture capabilities of spycraft to do
many more specialized things. In particular, external monitoring can also
provide generic alarming for dead or failing nodes and various diagnostics.

## More sip utilities

Having a good generic sip ping and endpoint status monitoring comes to mind.

