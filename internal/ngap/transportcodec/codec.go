// Package transportcodec validates NGAP PDUs and selects their SCTP stream.
package transportcodec

import (
	n2 "github.com/nycu-ucr/l25gc-n2-transport"
	"github.com/nycu-ucr/ngap"
)

func Classify(data []byte) (bool, error) {
	pdu, err := ngap.Decoder(data)
	if err != nil {
		return false, err
	}
	return n2.UEAssociated(pdu), nil
}
