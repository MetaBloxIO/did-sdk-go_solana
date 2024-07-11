# Project Usage

To use the MetaBlox DID SDK for Go, follow the steps below:

1. Install Go: Ensure that you have Go installed on your machine. You can download it from the official Go website

2. Download the MetaBlox DID SDK: Run the following command:

    ```
    go get github.com/MetaBloxIO/did-sdk-go-solana
    ```

3. Import the package: Import the DID SDK package in your Go code:

    ```go
    import "github.com/MetaBloxIO/did-sdk-go-solana"
    ```

# Example

1. Create an ED25519 Key Pair, which will be used for generating a DID and signing a Verifiable Document as an Issuer

    ```go
	pubKeyEd, privKeyEd, _ := ed25519.GenerateKey(rand.Reader)
    ```
2. Initialize the Issuer List and Contract Object

    ```go
	did.SetIssuerEDPrivateKey(&privKeyEd)
	did.InitIssuerDIDs([]string{did.SolanaChainName})
	did.InitBoundedContracts([]string{did.SolanaChainName})
    ```
3. Use the `did.CreateDID` function to create a new decentralized identifier (DID). Pass the Public Key and the chain name to the function

    ```go
	didEd := did.GenerateDIDString(&pubKeyEd, did.SolanaChainName)  // did:metablox:solana:<identifier>
    ```
4. Now create a Verifiable Credential (VC) for the DID
    
    ```go
    // For this example, create another ED25519 Key Pair as the Document holder
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
    ```

5. Finally, sign the VC with appropriate methods

    ```go
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
    ```

6. To verify the signed VC
    
    ```go
	valid, err := did.VerifyVC(vcEd)
	if err != nil || !valid {
		fmt.Printf("FAILED TO VERIFY VC: %s", err)
	}
    ```

7. To create the VP as the Presenter and to verify the VP, follow similar steps as for the VC

    ```go
	// Generating the DID Document for the Presenter
	_, presenterDocEd, _ := did.Resolve(didPresenter, nil, boundSol)
	// Create the VP using Presenter's ED25519 Private Key
	// Replace the "NONCE" according to your design, examples are the current Chain Block Height or Unix Timestamp
	vpEd, _ := did.CreatePresentation([]did.VerifiableCredential{*vcEd}, *presenterDocEd, &presenterPrivKey, "NONCE", did.Ed25519Sig)

	valid, err = did.VerifyVP(vpEd)
	if err != nil || !valid {
		fmt.Printf("FAILED TO VERIFY VP: %s", err)
	}
    ```

8. Use the JSON library to create a JSON version of the Verifiable Document

    ```go
	vcStr, _ := json.Marshal(vcEd)
    vpStr _ := json.Marshal(vpEd)
    ```