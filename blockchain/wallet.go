package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, pubKey := newKeyPair()
	wallet := &Wallet{private, pubKey}
	return wallet
}

func (w *Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPayload := append([]byte(version), pubKeyHash...)
	checksum := checksum(versionPayload)
	fullyPayload := append(versionPayload, checksum...)
	address := Base58Encode(fullyPayload)
	return address
}

func HashPubKey(pubKey []byte) []byte {
	publicSha256 := sha256.Sum256(pubKey)
	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSha256[:])
	if err != nil {
		log.Panicln(err)
	}
	return RIPEMD160Hasher.Sum(nil)
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	version := pubKeyHash[0]
	actualChecksum := pubKeyHash[len(pubKeyHash) - addressChecksumLen:]
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(targetChecksum, actualChecksum) == 0
}

func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])
	return secondSHA[:addressChecksumLen]
}

func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panicln(err)
	}
	pubKey := append(private.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}













