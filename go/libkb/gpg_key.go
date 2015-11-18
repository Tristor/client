// Copyright 2015 Keybase, Inc. All rights reserved. Use of
// this source code is governed by the included BSD license.

package libkb

import (
	"crypto/sha256"
	"errors"

	keybase1 "github.com/keybase/client/go/protocol"
	triplesec "github.com/keybase/go-triplesec"
)

// GPGKey is a shell around gpg cli commands that implements the
// GenericKey interface.
type GPGKey struct {
	fp  *PGPFingerprint
	kid keybase1.KID
	Contextified
}

// GPGKey implements the GenericKey interface.
var _ GenericKey = (*GPGKey)(nil)

func NewGPGKey(g *GlobalContext, fp *PGPFingerprint, kid keybase1.KID) *GPGKey {
	return &GPGKey{Contextified: NewContextified(g), fp: fp, kid: kid}
}

func (g *GPGKey) GetKID() keybase1.KID {
	return g.kid
}

func (g *GPGKey) GetFingerprintP() *PGPFingerprint {
	return g.fp
}

func (g *GPGKey) GetAlgoType() AlgoType {
	return KIDPGPBase
}

func (g *GPGKey) SignToString(msg []byte) (sig string, id keybase1.SigID, err error) {
	g.G().Log.Debug("GPGKey Signing %s", string(msg))
	sig, err = g.G().GetGpgClient().Sign(*g.fp, msg)
	if err != nil {
		return sig, id, err
	}

	// compute sig id:
	h := sha256.New()
	h.Write(msg)
	id, err = keybase1.SigIDFromSlice(h.Sum(nil))
	if err != nil {
		return sig, id, err
	}

	return sig, id, nil
}

func (g *GPGKey) VerifyStringAndExtract(sig string) (msg []byte, id keybase1.SigID, err error) {
	return msg, id, errors.New("VerifyStringAndExtract not implemented")
}

func (g *GPGKey) VerifyString(sig string, msg []byte) (id keybase1.SigID, err error) {
	return id, errors.New("VerifyString not implemented")
}

func (g *GPGKey) EncryptToString(plaintext []byte, sender GenericKey) (ciphertext string, err error) {
	return ciphertext, errors.New("EncryptToString not implemented")
}

func (g *GPGKey) DecryptFromString(ciphertext string) (msg []byte, sender keybase1.KID, err error) {
	return msg, sender, errors.New("DecryptFromString not implemented")
}

func (g *GPGKey) ToServerSKB(gc *GlobalContext, ts *triplesec.Cipher, gen PassphraseGeneration) (*SKB, error) {
	return nil, errors.New("ToServerSKB not implemented")
}

func (g *GPGKey) ToLksSKB(lks *LKSec) (*SKB, error) {
	return nil, errors.New("ToLksSKB not implemented")
}

func (g *GPGKey) VerboseDescription() string {
	return ""
}

func (g *GPGKey) CheckSecretKey() error {
	return nil
}

func (g *GPGKey) CanSign() bool {
	return true
}

func (g *GPGKey) CanEncrypt() bool {
	return false
}

func (g *GPGKey) CanDecrypt() bool {
	return false
}

func (g *GPGKey) HasSecretKey() bool {
	return true
}

func (g *GPGKey) Encode() (string, error) {
	return "", errors.New("Encode not implemented")
}