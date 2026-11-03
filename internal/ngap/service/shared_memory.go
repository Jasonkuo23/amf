package service

import (
	"context"
	"fmt"

	"github.com/nycu-ucr/amf/internal/logger"
	"github.com/nycu-ucr/amf/internal/ngap/transportcodec"
	n2 "github.com/nycu-ucr/l25gc-n2-transport"
)

var sharedEndpoint *n2.Endpoint
var sharedCancel context.CancelFunc

// RunSharedMemory uses the EAL already initialized by XIO. SCTP payloads still
// pass through the ordinary AMF NGAP dispatcher and connection lifecycle.
func RunSharedMemory(config *n2.Config, handler NGAPHandler) error {
	if config == nil {
		return fmt.Errorf("n2SharedMemory is required for sctp-shm")
	}
	cfg := *config
	cfg.Classify = transportcodec.Classify
	ep, err := n2.Open(cfg, true)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	sharedEndpoint = ep
	sharedCancel = cancel
	go func() {
		conn, err := ep.Accept(ctx)
		if err != nil {
			logger.NgapLog.Errorf("N2 shared-memory accept: %v", err)
			return
		}
		connections.Store(conn, conn)
		defer connections.Delete(conn)
		defer conn.Close()
		defer handler.HandleConnectionError(conn)
		buf := make([]byte, n2.MaxPDU)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				logger.NgapLog.Warnf("N2 association closed: %v; restart both endpoints to reconnect", err)
				return
			}
			// The handler may retain the PDU; copy only its length, not 256 KiB.
			handler.HandleMessage(conn, append([]byte(nil), buf[:n]...))
		}
	}()
	logger.NgapLog.Infof("N2 SCTP/IPv4 shared-memory listener ready: %s", cfg.Name)
	return nil
}
