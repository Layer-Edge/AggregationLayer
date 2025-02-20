package da

import (
    // "bytes"
    // "crypto/md5"
    // "encoding/binary"
    // "encoding/hex"
    "fmt"
    // "os/exec"
    // "strings"
)

type ZKProof struct {
}

// NO-OP for now
func (prf *ZKProof) GenerateProof(msg []byte) []byte {
    return msg
}

// MD5 sum for now
func (prf *ZKProof) GenerateAggregatedProof(msg string) string {
    /* struct {
        byte[2] length
        data
    }
        */
        merkleRoot, err := GetMerkleRoot(msg)
        if err != nil {
            fmt.Printf("Error generating Merkle root: %v\n", err)
            return ""
        }
        return merkleRoot
}

