package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"

	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const addressChecksumLen = 4
const walletFile = "wallet.json"

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// WalletData is used for JSON serialization
type WalletData struct {
	PrivateKeyD []byte `json:"private_key_d"`
	PrivateKeyX []byte `json:"private_key_x"`
	PrivateKeyY []byte `json:"private_key_y"`
	PublicKey   []byte `json:"public_key"`
}

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// Base58 alphabet
var b58Alphabet = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

// Base58Encode encodes a byte array to Base58
func Base58Encode(input []byte) []byte {
	var result []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(b58Alphabet)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		result = append(result, b58Alphabet[mod.Int64()])
	}

	// Add leading zeros
	if input[0] == 0x00 {
		result = append(result, b58Alphabet[0])
	}

	// Reverse the result
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Base58Decode decodes a Base58 string to a byte array
func Base58Decode(input []byte) []byte {
	result := big.NewInt(0)

	for _, b := range input {
		charIndex := -1
		for i, c := range b58Alphabet {
			if c == b {
				charIndex = i
				break
			}
		}
		if charIndex == -1 {
			log.Panic("Invalid character in Base58 string")
		}
		result.Mul(result, big.NewInt(58))
		result.Add(result, big.NewInt(int64(charIndex)))
	}

	decoded := result.Bytes()

	// Add leading zeros
	if input[0] == b58Alphabet[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

// NewKeyPair generates a new private/public key pair
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)

	fullPayload := append(versionedPayload, checksum...)
	address := Base58Encode(fullPayload)

	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// ValidateAddress validates wallet address
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	for i, b := range actualChecksum {
		if b != targetChecksum[i] {
			return false
		}
	}

	return true
}

// checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

// NewWallets creates Wallets and fills it from a file if it exists
func NewWallets() (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	err := wallets.LoadFromFile()

	return &wallets, err
}

// CreateWallet adds a Wallet to Wallets
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

// GetAddresses returns an array of addresses stored in the wallet file
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet returns a Wallet by its address
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from the file
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var walletsData map[string]WalletData
	err = json.Unmarshal(fileContent, &walletsData)
	if err != nil {
		log.Panic(err)
	}

	for address, walletData := range walletsData {
		// Reconstruct the private key
		curve := elliptic.P256()
		privateKey := &ecdsa.PrivateKey{
			PublicKey: ecdsa.PublicKey{
				Curve: curve,
				X:     new(big.Int).SetBytes(walletData.PrivateKeyX),
				Y:     new(big.Int).SetBytes(walletData.PrivateKeyY),
			},
			D: new(big.Int).SetBytes(walletData.PrivateKeyD),
		}

		wallet := &Wallet{
			PrivateKey: *privateKey,
			PublicKey:  walletData.PublicKey,
		}

		ws.Wallets[address] = wallet
	}

	return nil
}

// SaveToFile saves wallets to a file
func (ws Wallets) SaveToFile() {
	walletsData := make(map[string]WalletData)

	for address, wallet := range ws.Wallets {
		walletData := WalletData{
			PrivateKeyD: wallet.PrivateKey.D.Bytes(),
			PrivateKeyX: wallet.PrivateKey.PublicKey.X.Bytes(),
			PrivateKeyY: wallet.PrivateKey.PublicKey.Y.Bytes(),
			PublicKey:   wallet.PublicKey,
		}
		walletsData[address] = walletData
	}

	jsonData, err := json.MarshalIndent(walletsData, "", "  ")
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, jsonData, 0644)
	if err != nil {
		log.Panic(err)
	}
}

// HashPubKey from Wallet struct method - moved here for package access
func (w *Wallet) HashPubKey(pubKey []byte) []byte {
	return HashPubKey(pubKey)
}
