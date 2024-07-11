package demo

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha512"
	"fmt"

	"github.com/MetaBloxIO/did-sdk-go-solana"

	"github.com/multiformats/go-multibase"
)

func Demo() {
	pubKeyEd, privKeyEd, _ := ed25519.GenerateKey(rand.Reader)

	did.SetIssuerEDPrivateKey(&privKeyEd)
	did.InitIssuerDIDs([]string{did.SolanaChainName, did.EthereumChainName})
	did.InitBoundedContracts([]string{did.SolanaChainName, did.EthereumChainName})

	didEd := did.GenerateDIDString(&pubKeyEd, did.SolanaChainName) // did:metablox:solana:<identifier>

	presenterPubkey, presenterPrivKey, _ := ed25519.GenerateKey(rand.Reader)
	didPresenter := did.GenerateDIDString(&presenterPubkey, did.SolanaChainName)

	subjectData := did.MiningLicenseInfo{
		CredentialID: "1",
		Serial:       "1234567890abcdef",
		ID:           didPresenter,
		Name:         "John Doe",
		Model:        "FooBar",
	}

	boundSol, _ := did.GetBoundedContract(did.SolanaChainName)

	// Generating the DID Document for the Issuer
	_, issuerDocEd, _ := did.Resolve(didEd, nil, boundSol)

	// Create the VC using ED25519 Signature
	vcEd, _ := did.CreateVC(issuerDocEd, did.Ed25519Sig)

	vcEd.CredentialSubject = subjectData
	vcEd.Type = append(vcEd.Type, did.TypeMining)

	// Get the HASH for ED25519 signature
	vcBytes, _ := did.ConvertVCToJWTPayload(*vcEd)
	hashedVC256 := sha512.Sum512(vcBytes)
	ed25519Sig, _ := did.CreateEd25519JWSSignature(&privKeyEd, hashedVC256[:])
	// Now fillup the Proof and append it back to the VC
	ed25519Proof := vcEd.Proof.(did.Ed25519VCProof)
	ed25519Proof.JWSSignature = ed25519Sig
	pubBase58, _ := multibase.Encode(multibase.Base58BTC, pubKeyEd)
	ed25519Proof.PublicKeyMultibase = pubBase58
	vcEd.Proof = ed25519Proof

	valid, err := did.VerifyVC(vcEd)
	if err != nil || !valid {
		fmt.Printf("FAILED TO VERIFY VC: %s", err)
	}

	// Generating the DID Document for the Presenter
	_, presenterDocEd, _ := did.Resolve(didPresenter, nil, boundSol)
	// Create the VP using Presenter's ED25519 Private Key
	// You could replace the "NONCE" according to your design, examples are the current Chain Block Height or Unix Timestamp
	vpEd, _ := did.CreatePresentation([]did.VerifiableCredential{*vcEd}, *presenterDocEd, &presenterPrivKey, "NONCE", did.Ed25519Sig)

	valid, err = did.VerifyVP(vpEd)
	if err != nil || !valid {
		fmt.Printf("FAILED TO VERIFY VP: %s", err)
	}

	// Alter the VP and try to verify it again
	vpEd.Holder = didEd
	valid, err = did.VerifyVP(vpEd)
	if err != nil || !valid {
		fmt.Printf("FAILED TO VERIFY VP: %s", err)
	}
}
