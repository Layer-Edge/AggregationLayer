package da

import (
    "crypto/md5"
)

type ZKProof struct {
}

// NO-OP for now
func (prf *ZKProof) GenerateProof(msg []byte) []byte {
    return msg
}

// MD5 sum for now
func (prf *ZKProof) GenerateAggregatedProof(msg []byte) []byte {
    h := md5.New()
    return h.Sum(msg)
}
