package handlers

import (
	"fmt"

	"github.com/costinm/hbone"
)

// Original comments, using SNI routing and a separate reverse-role connection. Client->Server on different connection
// from Server->Client streams.
// H2R implements a reverse H2 connection:
// - Bob connects to the Gate's H2R port (or H2 port with ALPN=h2r), using mTLS,
// - gate authenticates bob, authorizes. Using the SNI in the TLS and cert.
// - Gate opens a client H2 connection on the accepted mTLS
// - Bob opens a H2 server connection on the dialed mTLS
// - Optional: Gate sends an 'init message', Bob responds. This may include additional meta.
//
// Either SNI or bob's cert SAN can be used to identify the cluster this belongs to.
//

// Alternative protocol using infra/POST (no changes to H2 or TLS, single connection, double mux):
// - Bob connects to regular hbone/POST port, starts a regular HBone H2 connection
// - Bob opens a stream (RoundTripStart)
// - Option 1: Gate opens an H2 client on the stream, uses stream headers to identify cluster
// - Option 2: Gate adds the stream to a 'listeners' list, forwards incoming stream directly
//   ( no encapsulation ).

// New implementation - extended H2 stack.

// WIP: RemoteForward is similar with ssh -R remotePort.
// Will use the H2R protocol to open a remote H2C connection
// attached to the Hbone remote server.
func RemoteForward(hb *hbone.HBone, hg, sn, ns string) *hbone.EndpointCon {
	attachC := &hbone.Cluster{
		ID:   "h2r-" + hg,
		Addr: sn + "." + ns + ":15009",
		SNI:  fmt.Sprintf("outbound_.%s_._.%s.%s.svc.cluster.local", "15009", sn, ns),
	}
	hb.AddService(attachC, &hbone.Endpoint{
		Address: attachC.Addr,
		Labels: map[string]string{
			"h2r": "1",
		},
	})

	attachE := &hbone.EndpointCon{
		Endpoint: &hbone.Endpoint{
			Address: attachC.Addr,
		},
		Cluster: attachC,
	}

	//go func() {
	//	_, err := DialH2R(context.Background(), attachE, hg)
	//	log.Println("H2R connected", hg, err)
	//}()

	return attachE
}

//// GetClientConn is called by http2.Transport, if Transport.RoundTrip is called (
//// for example used in a http.Client ). We are using the http2.ClientConn directly,
//// but this method may be needed if this library is used as a http client.
//func (hb *HBone) GetClientConn(req *http.Request, addr string) (*http2.ClientConn, error) {
//	c, err := hb.Cluster(req.Context(), addr)
//	if err != nil {
//		return nil, err
//	}
//
//	m, err := c.findMux(req.Context())
//	if err != nil {
//		return nil, err
//	}
//
//	return m.rt.(*http2.ClientConn), nil
//}
//
//func (hb *HBone) MarkDead(conn *http2.ClientConn) {
//	hb.m.Lock()
//	sni := hb.H2RConn[conn]
//
//	if sni != nil {
//		log.Println("H2RSNI: MarkDead ", sni)
//	}
//	delete(hb.H2RConn, conn)
//	hb.m.Unlock()
//
//}

//// HandleH2RConn takes a connection on the H2R port or on a stream and
//// implements a reverse connection.
//func (hb *HBone) HandlerH2RConn(conn net.Conn) {
//	conf := hb.Auth.GenerateTLSConfigServer()
//
//	tls := tls.Server(conn, conf)
//
//	err := nio.HandshakeTimeout(tls, hb.HandsahakeTimeout, conn)
//	if err != nil {
//		conn.Close()
//		return
//	}
//
//	// At this point we have the client identity, and we know it's in the trust domain and right CA.
//	// TODO: save the endpoint.
//
//	// TODO: construct the SNI header, save it in the map
//	// TODO: validate the trust domain, root cert, etc
//
//	sni := tls.ConnectionState().ServerName
//	if Debug {
//		log.Println("H2RSNI: accepted ", sni)
//	}
//	ctx := context.Background()
//
//	c, _ := hb.Cluster(ctx, sni)
//
//	// not blocking. Will write the 'preface' and start reading.
//	// When done, MarkDead on the conn pool in the transport is called.
//	rt, err := hb.h2t.NewClientConn(tls)
//	if err != nil {
//		conn.Close()
//		return
//	}
//
//	ec := &EndpointCon{
//		c:        c,
//		Endpoint: &Endpoint{Address: sni},
//		rt:       rt,
//	}
//	hb.m.Lock()
//	c.EndpointCon = append(c.EndpointCon, ec)
//	hb.m.Unlock()
//
//	hb.H2RConn[rt] = ec
//
//	// TODO: track the active connections in hb, for close purpose.
//}
