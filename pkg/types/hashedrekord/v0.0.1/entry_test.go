//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hashedrekord

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"reflect"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/sigstore/rekor/pkg/generated/models"
	x509r "github.com/sigstore/rekor/pkg/pki/x509"
	"github.com/sigstore/rekor/pkg/types"
	"github.com/sigstore/sigstore/pkg/signature"
	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNewEntryReturnType(t *testing.T) {
	entry := NewEntry()
	if reflect.TypeOf(entry) != reflect.ValueOf(&V001Entry{}).Type() {
		t.Errorf("invalid type returned from NewEntry: %T", entry)
	}
}

func TestCrossFieldValidation(t *testing.T) {
	type TestCase struct {
		caseDesc                  string
		entry                     V001Entry
		expectUnmarshalSuccess    bool
		expectCanonicalizeSuccess bool
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	keyBytes := pem.EncodeToMemory(&pem.Block{
		Bytes: der,
		Type:  "PUBLIC KEY",
	})

	// testing lack of support for ed25519
	invalidEdPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	invalidDer, err := x509.MarshalPKIXPublicKey(invalidEdPubKey)
	if err != nil {
		t.Fatal(err)
	}
	invalidKeyBytes := pem.EncodeToMemory(&pem.Block{
		Bytes: invalidDer,
		Type:  "PUBLIC KEY",
	})

	dataBytes := []byte("sign me!")
	h := sha256.Sum256(dataBytes)
	dataSHA := hex.EncodeToString(h[:])

	signer, _ := signature.LoadSigner(key, crypto.SHA256)
	sigBytes, _ := signer.SignMessage(bytes.NewReader(dataBytes))

	incorrectLengthHash := sha256.Sum224(dataBytes)
	incorrectLengthSHA := hex.EncodeToString(incorrectLengthHash[:])

	badHash := sha256.Sum256(keyBytes)
	badDataSHA := hex.EncodeToString(badHash[:])

	testCases := []TestCase{
		{
			caseDesc:               "empty obj",
			entry:                  V001Entry{},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature without url or content",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature without public key",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
					},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature with empty public key",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content:   sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{},
					},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature with ed25519 public key",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: invalidKeyBytes,
						},
					},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature without data",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: keyBytes,
						},
					},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature with empty data",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: keyBytes,
						},
					},
					Data: &models.HashedrekordV001SchemaData{},
				},
			},
			expectUnmarshalSuccess: false,
		},
		{
			caseDesc: "signature with hash",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: keyBytes,
						},
					},
					Data: &models.HashedrekordV001SchemaData{
						Hash: &models.HashedrekordV001SchemaDataHash{
							Value:     swag.String(dataSHA),
							Algorithm: swag.String(models.HashedrekordV001SchemaDataHashAlgorithmSha256),
						},
					},
				},
			},
			expectUnmarshalSuccess:    true,
			expectCanonicalizeSuccess: true,
		},
		{
			caseDesc: "signature with invalid sha length",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: keyBytes,
						},
					},
					Data: &models.HashedrekordV001SchemaData{
						Hash: &models.HashedrekordV001SchemaDataHash{
							Value:     swag.String(incorrectLengthSHA),
							Algorithm: swag.String(models.HashedrekordV001SchemaDataHashAlgorithmSha256),
						},
					},
				},
			},
			expectUnmarshalSuccess:    false,
			expectCanonicalizeSuccess: false,
		},
		{
			caseDesc: "signature with hash & invalid signature",
			entry: V001Entry{
				HashedRekordObj: models.HashedrekordV001Schema{
					Signature: &models.HashedrekordV001SchemaSignature{
						Content: sigBytes,
						PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{
							Content: keyBytes,
						},
					},
					Data: &models.HashedrekordV001SchemaData{
						Hash: &models.HashedrekordV001SchemaDataHash{
							Value:     swag.String(badDataSHA),
							Algorithm: swag.String(models.HashedrekordV001SchemaDataHashAlgorithmSha256),
						},
					},
				},
			},
			expectUnmarshalSuccess:    false,
			expectCanonicalizeSuccess: false,
		},
	}

	for _, tc := range testCases {
		if _, _, err := tc.entry.validate(); (err == nil) != tc.expectUnmarshalSuccess {
			t.Errorf("unexpected result in '%v': %v", tc.caseDesc, err)
		}

		v := &V001Entry{}
		r := models.Hashedrekord{
			APIVersion: swag.String(tc.entry.APIVersion()),
			Spec:       tc.entry.HashedRekordObj,
		}

		unmarshalAndValidate := func() error {
			if err := v.Unmarshal(&r); err != nil {
				return err
			}
			if _, _, err := v.validate(); err != nil {
				return err
			}
			return nil
		}

		if err := unmarshalAndValidate(); (err == nil) != tc.expectUnmarshalSuccess {
			t.Errorf("unexpected result in '%v': %v", tc.caseDesc, err)
		}

		b, err := v.Canonicalize(context.TODO())
		if (err == nil) != tc.expectCanonicalizeSuccess {
			t.Errorf("unexpected result from Canonicalize for '%v': %v", tc.caseDesc, err)
		} else if err != nil {
			if _, ok := err.(types.ValidationError); !ok {
				t.Errorf("canonicalize returned an unexpected error that isn't of type types.ValidationError: %v", err)
			}
		}
		if b != nil {
			pe, err := models.UnmarshalProposedEntry(bytes.NewReader(b), runtime.JSONConsumer())
			if err != nil {
				t.Errorf("unexpected err from Unmarshalling canonicalized entry for '%v': %v", tc.caseDesc, err)
			}
			if _, err := types.UnmarshalEntry(pe); err != nil {
				t.Errorf("unexpected err from type-specific unmarshalling for '%v': %v", tc.caseDesc, err)
			}
		}
	}
}

func hexHash(b []byte) string {
	h := sha256.Sum256([]byte(b))
	return hex.EncodeToString(h[:])
}

func TestV001Entry_IndexKeys(t *testing.T) {
	pub, cert, priv := testKeyAndCert(t)

	data := "my random data"
	h := sha256.Sum256([]byte(data))
	sig, err := ecdsa.SignASN1(rand.Reader, priv, h[:])
	if err != nil {
		t.Fatal(err)
	}

	hashStr := hex.EncodeToString(h[:])
	hashIndexKey := "sha256:" + hashStr
	// Base entry template
	v := V001Entry{
		HashedRekordObj: models.HashedrekordV001Schema{
			Data: &models.HashedrekordV001SchemaData{
				Hash: &models.HashedrekordV001SchemaDataHash{
					Algorithm: swag.String("sha256"),
					Value:     swag.String(hashStr),
				},
			},
			Signature: &models.HashedrekordV001SchemaSignature{
				Content:   strfmt.Base64(sig),
				PublicKey: &models.HashedrekordV001SchemaSignaturePublicKey{},
			},
		},
	}

	// Test with a public key and a cert

	// For the public key, we should have the key and the hash.
	t.Run("public key", func(t *testing.T) {
		v.HashedRekordObj.Signature.PublicKey.Content = strfmt.Base64(pub)

		k, err := v.IndexKeys()
		if err != nil {
			t.Fatal(err)
		}
		keys := map[string]struct{}{}
		for _, key := range k {
			keys[key] = struct{}{}
		}

		if _, ok := keys[hashIndexKey]; !ok {
			t.Errorf("missing hash index entry %s, got %v", hashIndexKey, keys)
		}
		want := hexHash(pub)
		if _, ok := keys[want]; !ok {
			t.Errorf("missing key index entry %s, got %v", want, keys)
		}
	})

	// For the public key, we should have the key and the hash.
	t.Run("cert", func(t *testing.T) {
		v.HashedRekordObj.Signature.PublicKey.Content = strfmt.Base64(cert)

		k, err := v.IndexKeys()
		if err != nil {
			t.Fatal(err)
		}
		keys := map[string]struct{}{}
		for _, key := range k {
			keys[key] = struct{}{}
		}

		if _, ok := keys[hashIndexKey]; !ok {
			t.Errorf("missing hash index entry for public key test, got %v", keys)
		}
		if _, ok := keys[hexHash(cert)]; !ok {
			t.Errorf("missing key index entry for public key test, got %v", keys)
		}
	})

}

func testKeyAndCert(t *testing.T) ([]byte, []byte, *ecdsa.PrivateKey) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	pub := pem.EncodeToMemory(&pem.Block{
		Bytes: der,
		Type:  "PUBLIC KEY",
	})

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Names: []pkix.AttributeTypeAndValue{
				{
					Type:  x509r.EmailAddressOID,
					Value: "foo@bar.com",
				},
			},
		},
	}
	cb, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	certPem := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cb,
	})

	return pub, certPem, priv
}
